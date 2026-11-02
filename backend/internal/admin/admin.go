package admin

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

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	var orders, customers, restaurants, partners int
	var gmv float64
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM orders").Scan(&orders)
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users").Scan(&customers)
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM restaurants").Scan(&restaurants)
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM delivery_partners").Scan(&partners)
	h.pool.QueryRow(r.Context(), "SELECT COALESCE(SUM(total),0) FROM orders").Scan(&gmv)
	writeJSON(w, 200, map[string]any{
		"total_orders": orders, "total_customers": customers,
		"total_restaurants": restaurants, "total_delivery_partners": partners,
		"gross_order_value": gmv,
	})
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, _ := h.pool.Query(r.Context(), "SELECT id, phone, email, status, created_at FROM users ORDER BY created_at DESC LIMIT 100")
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id string
		var phone, email, status *string
		var created interface{}
		rows.Scan(&id, &phone, &email, &status, &created)
		out = append(out, map[string]any{"id": id, "phone": phone, "email": email, "status": status, "created_at": created})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	rows, _ := h.pool.Query(r.Context(), "SELECT id, name, status, avg_rating, total_ratings, created_at FROM restaurants ORDER BY created_at DESC LIMIT 100")
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, name, status string
		var rating float64
		var total int
		var created interface{}
		rows.Scan(&id, &name, &status, &rating, &total, &created)
		out = append(out, map[string]any{"id": id, "name": name, "status": status, "avg_rating": rating, "total_ratings": total, "created_at": created})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	rows, _ := h.pool.Query(r.Context(), "SELECT id, order_number, status, payment_status, total, placed_at FROM orders ORDER BY placed_at DESC LIMIT 100")
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