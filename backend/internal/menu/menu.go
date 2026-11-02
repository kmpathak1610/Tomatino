package menu

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

func (h *Handler) ListByRestaurant(w http.ResponseWriter, r *http.Request) {
	rid := r.URL.Query().Get("restaurant_id")
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, name, description, base_price, is_veg, is_available, is_bestseller, prep_time_min, image_url
		FROM menu_items WHERE restaurant_id=$1 ORDER BY created_at`, rid)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, name string
		var desc, img *string
		var price float64
		var isVeg, avail, best bool
		var prep int
		rows.Scan(&id, &name, &desc, &price, &isVeg, &avail, &best, &prep, &img)
		out = append(out, map[string]any{
			"id": id, "name": name, "description": desc, "base_price": price,
			"is_veg": isVeg, "is_available": avail, "is_bestseller": best,
			"prep_time_min": prep, "image_url": img,
		})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestaurantID string  `json:"restaurant_id"`
		CategoryID   *string `json:"category_id"`
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		BasePrice    float64 `json:"base_price"`
		IsVeg        bool    `json:"is_veg"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	id := uuid.New().String()
	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO menu_items(id,restaurant_id,category_id,name,description,base_price,is_veg)
		VALUES($1,$2,$3,$4,$5,$6,$7)`,
		id, req.RestaurantID, req.CategoryID, req.Name, req.Description, req.BasePrice, req.IsVeg)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}