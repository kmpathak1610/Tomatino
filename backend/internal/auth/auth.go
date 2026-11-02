package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	pool   *pgxpool.Pool
	secret []byte
}

func NewService(pool *pgxpool.Pool) *Service {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-change-me"
	}
	return &Service{pool: pool, secret: []byte(s)}
}

func (s *Service) RegisterRoutes(mux interface{ HandleFunc(string, func(http.ResponseWriter, *http.Request)) }) {
	// placeholder: routes wired via chi in main
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// Dev OTP: store in memory-ish via DB? We'll just accept any 6-digit OTP for dev.
type otpReq struct {
	Phone string `json:"phone"`
}
type verifyReq struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

func (s *Service) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req otpReq
	json.NewDecoder(r.Body).Decode(&req)
	// dev adapter: no real SMS
	writeJSON(w, 200, map[string]any{"message": "OTP sent (dev: use 123456)", "phone": req.Phone})
}

func (s *Service) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyReq
	json.NewDecoder(r.Body).Decode(&req)
	if req.OTP != "123456" {
		writeJSON(w, 400, map[string]string{"error": "invalid OTP"})
		return
	}
	ctx := r.Context()
	// upsert user
	var userID string
	err := s.pool.QueryRow(ctx, "SELECT id FROM users WHERE phone=$1", req.Phone).Scan(&userID)
	if err != nil {
		userID = uuid.New().String()
		_, err = s.pool.Exec(ctx, "INSERT INTO users(id,phone,status) VALUES($1,$2,'ACTIVE')", userID, req.Phone)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		_, _ = s.pool.Exec(ctx, "INSERT INTO user_profiles(user_id,full_name) VALUES($1,$2) ON CONFLICT DO NOTHING", userID, "")
		_, _ = s.pool.Exec(ctx, "INSERT INTO user_roles(user_id,role) VALUES($1,'CUSTOMER') ON CONFLICT DO NOTHING", userID)
	}
	// issue JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"phone": req.Phone,
		"role":  "CUSTOMER",
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	ts, _ := token.SignedString(s.secret)
	writeJSON(w, 200, map[string]any{"access_token": ts, "user_id": userID})
}

func (s *Service) StaffLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	var id, hash string
	var role string
	err := s.pool.QueryRow(r.Context(), `
		SELECT u.id, u.password_hash, ur.role FROM users u
		JOIN user_roles ur ON ur.user_id=u.id
		WHERE u.email=$1 LIMIT 1`, req.Email).Scan(&id, &hash, &role)
	if err != nil {
		writeJSON(w, 401, map[string]string{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		writeJSON(w, 401, map[string]string{"error": "invalid credentials"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  id,
		"role": role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	ts, _ := token.SignedString(s.secret)
	writeJSON(w, 200, map[string]any{"access_token": ts, "user_id": id, "role": role})
}

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			writeJSON(w, 401, map[string]string{"error": "missing token"})
			return
		}
		tokenStr := auth
		if len(auth) > 7 && auth[:7] == "Bearer " {
			tokenStr = auth[7:]
		}
		tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return s.secret, nil })
		if err != nil || !tok.Valid {
			writeJSON(w, 401, map[string]string{"error": "invalid token"})
			return
		}
		claims := tok.Claims.(jwt.MapClaims)
		r.Header.Set("X-User-ID", claims["sub"].(string))
		r.Header.Set("X-User-Role", claims["role"].(string))
		next.ServeHTTP(w, r)
	})
}

func (s *Service) UserID(r *http.Request) string { return r.Header.Get("X-User-ID") }
func (s *Service) UserRole(r *http.Request) string { return r.Header.Get("X-User-Role") }