package cart

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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var cartID, restID *string
	err := h.pool.QueryRow(r.Context(), "SELECT id, restaurant_id FROM carts WHERE user_id=$1", userID).Scan(&cartID, &restID)
	if err != nil {
		writeJSON(w, 200, map[string]any{"items": []any{}, "restaurant_id": nil})
		return
	}
	rows, _ := h.pool.Query(r.Context(), `
		SELECT ci.id, ci.menu_item_id, ci.variant_id, ci.quantity, ci.unit_price, ci.special_instructions,
		       mi.name, mi.is_veg
		FROM cart_items ci JOIN menu_items mi ON mi.id=ci.menu_item_id
		WHERE ci.cart_id=$1`, *cartID)
	defer rows.Close()
	var items []map[string]any
	for rows.Next() {
		var id, mid string
		var vid, instr *string
		var qty int
		var price float64
		var name string
		var veg bool
		rows.Scan(&id, &mid, &vid, &qty, &price, &instr, &name, &veg)
		items = append(items, map[string]any{
			"id": id, "menu_item_id": mid, "variant_id": vid, "quantity": qty,
			"unit_price": price, "special_instructions": instr, "name": name, "is_veg": veg,
		})
	}
	if items == nil {
		items = []map[string]any{}
	}
	writeJSON(w, 200, map[string]any{"cart_id": cartID, "restaurant_id": restID, "items": items})
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		RestaurantID string  `json:"restaurant_id"`
		BranchID     string  `json:"branch_id"`
		MenuItemID   string  `json:"menu_item_id"`
		Quantity     int     `json:"quantity"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Quantity < 1 {
		req.Quantity = 1
	}
	// fetch price
	var price float64
	var avail bool
	err := h.pool.QueryRow(r.Context(), "SELECT base_price, is_available FROM menu_items WHERE id=$1", req.MenuItemID).Scan(&price, &avail)
	if err != nil || !avail {
		writeJSON(w, 400, map[string]string{"error": "item unavailable"})
		return
	}
	// get or create cart
	var cartID string
	var existingRest *string
	err = h.pool.QueryRow(r.Context(), "SELECT id, restaurant_id FROM carts WHERE user_id=$1", userID).Scan(&cartID, &existingRest)
	if err != nil {
		cartID = uuid.New().String()
		_, _ = h.pool.Exec(r.Context(), "INSERT INTO carts(id,user_id,restaurant_id,branch_id) VALUES($1,$2,$3,$4)", cartID, userID, req.RestaurantID, req.BranchID)
	} else {
		if existingRest != nil && *existingRest != req.RestaurantID {
			writeJSON(w, 400, map[string]string{"error": "cart contains items from another restaurant; clear cart first"})
			return
		}
		_, _ = h.pool.Exec(r.Context(), "UPDATE carts SET restaurant_id=$1, branch_id=$2, updated_at=now() WHERE id=$3", req.RestaurantID, req.BranchID, cartID)
	}
	// upsert item
	var existingItem string
	err = h.pool.QueryRow(r.Context(), "SELECT id FROM cart_items WHERE cart_id=$1 AND menu_item_id=$2", cartID, req.MenuItemID).Scan(&existingItem)
	if err == nil {
		_, _ = h.pool.Exec(r.Context(), "UPDATE cart_items SET quantity=quantity+$1 WHERE id=$2", req.Quantity, existingItem)
	} else {
		_, _ = h.pool.Exec(r.Context(), "INSERT INTO cart_items(id,cart_id,menu_item_id,quantity,unit_price) VALUES($1,$2,$3,$4,$5)", uuid.New().String(), cartID, req.MenuItemID, req.Quantity, price)
	}
	writeJSON(w, 200, map[string]string{"status": "added"})
}

func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = h.pool.Exec(r.Context(), "DELETE FROM cart_items WHERE id=$1", id)
	writeJSON(w, 200, map[string]string{"status": "removed"})
}

func (h *Handler) Clear(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var cartID string
	h.pool.QueryRow(r.Context(), "SELECT id FROM carts WHERE user_id=$1", userID).Scan(&cartID)
	if cartID != "" {
		_, _ = h.pool.Exec(r.Context(), "DELETE FROM cart_items WHERE cart_id=$1", cartID)
		_, _ = h.pool.Exec(r.Context(), "DELETE FROM carts WHERE id=$1", cartID)
	}
	writeJSON(w, 200, map[string]string{"status": "cleared"})
}