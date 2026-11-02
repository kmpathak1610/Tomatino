package reviews

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		OrderID      string `json:"order_id"`
		RestaurantID string `json:"restaurant_id"`
		Rating       int    `json:"rating"`
		Comment      string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Rating < 1 || req.Rating > 5 {
		writeJSON(w, 400, map[string]string{"error": "rating 1-5"})
		return
	}
	id := uuid.New().String()
	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO restaurant_reviews(id,order_id,restaurant_id,user_id,rating,comment)
		VALUES($1,$2,$3,$4,$5,$6)`, id, req.OrderID, req.RestaurantID, userID, req.Rating, req.Comment)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	// update avg rating
	_, _ = h.pool.Exec(r.Context(), `
		UPDATE restaurants SET
		  avg_rating = (SELECT COALESCE(AVG(rating),0) FROM restaurant_reviews WHERE restaurant_id=$1),
		  total_ratings = (SELECT COUNT(*) FROM restaurant_reviews WHERE restaurant_id=$1),
		  updated_at = now()
		WHERE id=$1`, req.RestaurantID)
	writeJSON(w, 201, map[string]any{"id": id})
}

func (h *Handler) ListByRestaurant(w http.ResponseWriter, r *http.Request) {
	rid := r.URL.Query().Get("restaurant_id")
	rows, _ := h.pool.Query(r.Context(), `
		SELECT id, user_id, rating, comment, created_at FROM restaurant_reviews
		WHERE restaurant_id=$1 ORDER BY created_at DESC LIMIT 50`, rid)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, uid string
		var rating int
		var comment *string
		var created interface{}
		rows.Scan(&id, &uid, &rating, &comment, &created)
		out = append(out, map[string]any{"id": id, "user_id": uid, "rating": rating, "comment": comment, "created_at": created})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}