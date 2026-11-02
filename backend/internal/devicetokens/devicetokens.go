package devicetokens

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Platform == "" {
		req.Platform = "WEB"
	}
	id := uuid.New().String()
	_, err := h.pool.Exec(r.Context(), "INSERT INTO device_tokens(id,user_id,token,platform) VALUES($1,$2,$3,$4) ON CONFLICT(token) DO NOTHING", id, userID, req.Token, req.Platform)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}