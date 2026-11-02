package orders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

var validTransitions = map[string][]string{
	"PENDING_RESTAURANT_ACCEPTANCE": {"ACCEPTED", "REJECTED", "CANCELLED"},
	"ACCEPTED":                      {"PREPARING", "CANCELLED"},
	"PREPARING":                     {"READY_FOR_PICKUP", "CANCELLED"},
	"READY_FOR_PICKUP":              {"PICKED_UP"},
	"PICKED_UP":                     {"OUT_FOR_DELIVERY"},
	"OUT_FOR_DELIVERY":              {"DELIVERED"},
}

func canTransition(from, to string) bool {
	for _, t := range validTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// Create order from cart (simplified)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		RestaurantID string `json:"restaurant_id"`
		BranchID     string `json:"branch_id"`
		AddressID    string `json:"address_id"`
		PaymentMethod string `json:"payment_method"`
		Items        []struct {
			MenuItemID string  `json:"menu_item_id"`
			Quantity   int     `json:"quantity"`
			UnitPrice  float64 `json:"unit_price"`
		} `json:"items"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// backend pricing: fetch actual prices
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer tx.Rollback(r.Context())

	var subtotal float64
	type item struct {
		id    string
		name  string
		price float64
		qty   int
		veg   bool
	}
	var items []item
	for _, it := range req.Items {
		var name string
		var price float64
		var veg bool
		err := tx.QueryRow(r.Context(), "SELECT name, base_price, is_veg FROM menu_items WHERE id=$1 AND is_available=TRUE", it.MenuItemID).Scan(&name, &price, &veg)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "item unavailable: " + it.MenuItemID})
			return
		}
		items = append(items, item{id: it.MenuItemID, name: name, price: price, qty: it.Quantity, veg: veg})
		subtotal += price * float64(it.Quantity)
	}
	deliveryFee := 30.0
	tax := subtotal * 0.05
	total := subtotal + deliveryFee + tax

	// resolve a valid branch for the restaurant
	branchID := req.BranchID
	var validBranch string
	if err := tx.QueryRow(r.Context(), "SELECT id FROM restaurant_branches WHERE restaurant_id=$1 AND is_active=TRUE LIMIT 1", req.RestaurantID).Scan(&validBranch); err == nil {
		branchID = validBranch
	}
	if branchID == "" {
		writeJSON(w, 400, map[string]string{"error": "no active branch for restaurant"})
		return
	}
	var addressID any
	if req.AddressID != "" {
		addressID = req.AddressID
	}

	orderID := uuid.New().String()
	orderNumber := fmt.Sprintf("ORD-%d", time.Now().Unix())
	paymentStatus := "PENDING"
	if req.PaymentMethod == "COD" {
		paymentStatus = "COD_PENDING"
	}
	_, err = tx.Exec(r.Context(), `
		INSERT INTO orders(id,order_number,user_id,restaurant_id,branch_id,address_id,status,payment_status,payment_method,subtotal,delivery_fee,tax,total,placed_at)
		VALUES($1,$2,$3,$4,$5,$6,'PENDING_RESTAURANT_ACCEPTANCE',$7,$8,$9,$10,$11,$12,now())`,
		orderID, orderNumber, userID, req.RestaurantID, branchID, addressID, paymentStatus, req.PaymentMethod, subtotal, deliveryFee, tax, total)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	for _, it := range items {
		oi := uuid.New().String()
		_, err = tx.Exec(r.Context(), `
			INSERT INTO order_items(id,order_id,menu_item_id,name_snapshot,unit_price,quantity,line_total,is_veg)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
			oi, orderID, it.id, it.name, it.price, it.qty, it.price*float64(it.qty), it.veg)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
	}
	// status history
	_, _ = tx.Exec(r.Context(), "INSERT INTO order_status_history(order_id,to_status,changed_by) VALUES($1,'PENDING_RESTAURANT_ACCEPTANCE', $2)", orderID, userID)
	tx.Commit(r.Context())

	writeJSON(w, 201, map[string]any{"order_id": orderID, "order_number": orderNumber, "total": total, "payment_status": paymentStatus})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	// fetch current
	var current string
	err := h.pool.QueryRow(r.Context(), "SELECT status FROM orders WHERE id=$1", id).Scan(&current)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	if !canTransition(current, req.Status) {
		writeJSON(w, 400, map[string]string{"error": "invalid transition"})
		return
	}
	actor := r.Header.Get("X-User-ID")
	_, err = h.pool.Exec(r.Context(), "UPDATE orders SET status=$1, updated_at=now() WHERE id=$2", req.Status, id)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	_, _ = h.pool.Exec(r.Context(), "INSERT INTO order_status_history(order_id,from_status,to_status,changed_by,note) VALUES($1,$2,$3,$4,$5)", id, current, req.Status, actor, req.Note)
	writeJSON(w, 200, map[string]string{"status": req.Status})
}

func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, order_number, status, payment_status, total, placed_at
		FROM orders WHERE user_id=$1 ORDER BY placed_at DESC LIMIT 50`, userID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, num, status, pstatus string
		var total float64
		var placed time.Time
		rows.Scan(&id, &num, &status, &pstatus, &total, &placed)
		out = append(out, map[string]any{"id": id, "order_number": num, "status": status, "payment_status": pstatus, "total": total, "placed_at": placed})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var num, status, pstatus, method string
	var total, subtotal float64
	err := h.pool.QueryRow(r.Context(), `
		SELECT order_number, status, payment_status, payment_method, subtotal, total FROM orders WHERE id=$1`, id).
		Scan(&num, &status, &pstatus, &method, &subtotal, &total)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "order_number": num, "status": status, "payment_status": pstatus, "payment_method": method, "subtotal": subtotal, "total": total})
}