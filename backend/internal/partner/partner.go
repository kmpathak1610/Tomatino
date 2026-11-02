package partner

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

func (h *Handler) myRestaurantIDs(r *http.Request) []string {
	userID := r.Header.Get("X-User-ID")
	rows, _ := h.pool.Query(r.Context(), `
		SELECT id FROM restaurants WHERE owner_user_id=$1
		UNION
		SELECT restaurant_id FROM restaurant_staff WHERE user_id=$1`, userID)
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids
}

func (h *Handler) Restaurant(w http.ResponseWriter, r *http.Request) {
	ids := h.myRestaurantIDs(r)
	if len(ids) == 0 {
		writeJSON(w, 404, map[string]string{"error": "no restaurant"})
		return
	}
	var id, name, status string
	var rating float64
	var total int
	h.pool.QueryRow(r.Context(), "SELECT id,name,status,avg_rating,total_ratings FROM restaurants WHERE id=$1", ids[0]).Scan(&id, &name, &status, &rating, &total)
	writeJSON(w, 200, map[string]any{"id": id, "name": name, "status": status, "avg_rating": rating, "total_ratings": total})
}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	ids := h.myRestaurantIDs(r)
	if len(ids) == 0 {
		writeJSON(w, 200, []any{})
		return
	}
	rows, _ := h.pool.Query(r.Context(), `
		SELECT id, order_number, status, payment_status, total, placed_at
		FROM orders WHERE restaurant_id = ANY($1) ORDER BY placed_at DESC LIMIT 100`, ids)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, num, status, pstatus string
		var total float64
		var placed interface{}
		rows.Scan(&id, &num, &status, &pstatus, &total, &placed)
		out = append(out, map[string]any{"id": id, "order_number": num, "status": status, "payment_status": pstatus, "total": total, "placed_at": placed})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	ids := h.myRestaurantIDs(r)
	if len(ids) == 0 {
		writeJSON(w, 200, map[string]any{"orders": 0, "revenue": 0})
		return
	}
	var orders int
	var revenue float64
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*), COALESCE(SUM(total),0) FROM orders WHERE restaurant_id = ANY($1)", ids).Scan(&orders, &revenue)
	var pending int
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM orders WHERE restaurant_id = ANY($1) AND status='PENDING_RESTAURANT_ACCEPTANCE'", ids).Scan(&pending)
	writeJSON(w, 200, map[string]any{"orders": orders, "revenue": revenue, "pending": pending})
}

func (h *Handler) Menu(w http.ResponseWriter, r *http.Request) {
	ids := h.myRestaurantIDs(r)
	if len(ids) == 0 {
		writeJSON(w, 200, []any{})
		return
	}
	rows, _ := h.pool.Query(r.Context(), "SELECT id,name,base_price,is_veg,is_available,is_bestseller FROM menu_items WHERE restaurant_id = ANY($1) ORDER BY name", ids)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, name string
		var price float64
		var veg, avail, best bool
		rows.Scan(&id, &name, &price, &veg, &avail, &best)
		out = append(out, map[string]any{"id": id, "name": name, "base_price": price, "is_veg": veg, "is_available": avail, "is_bestseller": best})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) ToggleAvailability(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Available bool `json:"available"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_, _ = h.pool.Exec(r.Context(), "UPDATE menu_items SET is_available=$1, updated_at=now() WHERE id=$2", req.Available, id)
	writeJSON(w, 200, map[string]string{"status": "ok"})
}