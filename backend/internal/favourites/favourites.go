package favourites

import (
	"encoding/json"
	"net/http"

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

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		RestaurantID string `json:"restaurant_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_, _ = h.pool.Exec(r.Context(), "INSERT INTO favourites(user_id,restaurant_id) VALUES($1,$2) ON CONFLICT DO NOTHING", userID, req.RestaurantID)
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	rid := r.PathValue("restaurant_id")
	_, _ = h.pool.Exec(r.Context(), "DELETE FROM favourites WHERE user_id=$1 AND restaurant_id=$2", userID, rid)
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	rows, _ := h.pool.Query(r.Context(), "SELECT restaurant_id, created_at FROM favourites WHERE user_id=$1", userID)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var rid string
		var created interface{}
		rows.Scan(&rid, &created)
		out = append(out, map[string]any{"restaurant_id": rid, "created_at": created})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}