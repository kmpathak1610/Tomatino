package delivery

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

func (h *Handler) partnerID(r *http.Request) string {
	userID := r.Header.Get("X-User-ID")
	var id string
	h.pool.QueryRow(r.Context(), "SELECT id FROM delivery_partners WHERE user_id=$1", userID).Scan(&id)
	return id
}

func (h *Handler) MyAssignments(w http.ResponseWriter, r *http.Request) {
	pid := h.partnerID(r)
	rows, _ := h.pool.Query(r.Context(), `
		SELECT da.id, da.order_id, da.status, da.delivery_otp, o.order_number, o.status AS order_status, o.total
		FROM delivery_assignments da JOIN orders o ON o.id=da.order_id
		WHERE da.delivery_partner_id=$1 ORDER BY da.assigned_at DESC LIMIT 50`, pid)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, oid, status, otp, num, ostatus string
		var total float64
		rows.Scan(&id, &oid, &status, &otp, &num, &ostatus, &total)
		out = append(out, map[string]any{"id": id, "order_id": oid, "status": status, "order_number": num, "order_status": ostatus, "total": total})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) Available(w http.ResponseWriter, r *http.Request) {
	rows, _ := h.pool.Query(r.Context(), `
		SELECT o.id, o.order_number, o.total, o.status
		FROM orders o
		LEFT JOIN delivery_assignments da ON da.order_id=o.id
		WHERE da.id IS NULL AND o.status IN ('READY_FOR_PICKUP','PREPARING','ACCEPTED')
		ORDER BY o.placed_at DESC LIMIT 50`)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, num, status string
		var total float64
		rows.Scan(&id, &num, &total, &status)
		out = append(out, map[string]any{"order_id": id, "order_number": num, "total": total, "status": status})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		VehicleType string `json:"vehicle_type"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	id := uuid.New().String()
	_, err := h.pool.Exec(r.Context(), "INSERT INTO delivery_partners(id,user_id,vehicle_type,status) VALUES($1,$2,$3,'PENDING')", id, userID, req.VehicleType)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	_, _ = h.pool.Exec(r.Context(), "INSERT INTO user_roles(user_id,role) VALUES($1,'DELIVERY_PARTNER') ON CONFLICT DO NOTHING", userID)
	writeJSON(w, 201, map[string]any{"id": id, "status": "PENDING"})
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = h.pool.Exec(r.Context(), "UPDATE delivery_partners SET status='APPROVED' WHERE id=$1", id)
	writeJSON(w, 200, map[string]string{"status": "APPROVED"})
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID          string `json:"order_id"`
		DeliveryPartnerID string `json:"delivery_partner_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	id := uuid.New().String()
	otp := "1234"
	_, err := h.pool.Exec(r.Context(), "INSERT INTO delivery_assignments(id,order_id,delivery_partner_id,delivery_otp) VALUES($1,$2,$3,$4)", id, req.OrderID, req.DeliveryPartnerID, otp)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	_, _ = h.pool.Exec(r.Context(), "UPDATE orders SET delivery_partner_id=$1 WHERE id=$2", req.DeliveryPartnerID, req.OrderID)
	writeJSON(w, 201, map[string]any{"id": id, "delivery_otp": otp})
}

func (h *Handler) Claim(w http.ResponseWriter, r *http.Request) {
	pid := h.partnerID(r)
	orderID := r.PathValue("order_id")
	id := uuid.New().String()
	otp := "1234"
	_, err := h.pool.Exec(r.Context(), "INSERT INTO delivery_assignments(id,order_id,delivery_partner_id,status,delivery_otp) VALUES($1,$2,$3,'ASSIGNED',$4)", id, orderID, pid, otp)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	_, _ = h.pool.Exec(r.Context(), "UPDATE orders SET delivery_partner_id=$1 WHERE id=$2", pid, orderID)
	writeJSON(w, 201, map[string]any{"id": id})
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = h.pool.Exec(r.Context(), "UPDATE delivery_assignments SET status='ACCEPTED', accepted_at=now() WHERE id=$1", id)
	writeJSON(w, 200, map[string]string{"status": "ACCEPTED"})
}

func (h *Handler) Pickup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var orderID string
	h.pool.QueryRow(r.Context(), "SELECT order_id FROM delivery_assignments WHERE id=$1", id).Scan(&orderID)
	_, _ = h.pool.Exec(r.Context(), "UPDATE delivery_assignments SET status='PICKED_UP', picked_up_at=now() WHERE id=$1", id)
	_, _ = h.pool.Exec(r.Context(), "UPDATE orders SET status='PICKED_UP', picked_up_at=now(), updated_at=now() WHERE id=$1 AND status='READY_FOR_PICKUP'", orderID)
	_, _ = h.pool.Exec(r.Context(), "INSERT INTO order_status_history(order_id,from_status,to_status,note) VALUES($1,'READY_FOR_PICKUP','PICKED_UP','picked up by rider')", orderID)
	writeJSON(w, 200, map[string]string{"status": "PICKED_UP"})
}

func (h *Handler) OutForDelivery(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var orderID string
	h.pool.QueryRow(r.Context(), "SELECT order_id FROM delivery_assignments WHERE id=$1", id).Scan(&orderID)
	_, _ = h.pool.Exec(r.Context(), "UPDATE delivery_assignments SET status='OUT_FOR_DELIVERY' WHERE id=$1", id)
	_, _ = h.pool.Exec(r.Context(), "UPDATE orders SET status='OUT_FOR_DELIVERY', updated_at=now() WHERE id=$1 AND status='PICKED_UP'", orderID)
	_, _ = h.pool.Exec(r.Context(), "INSERT INTO order_status_history(order_id,from_status,to_status) VALUES($1,'PICKED_UP','OUT_FOR_DELIVERY')", orderID)
	writeJSON(w, 200, map[string]string{"status": "OUT_FOR_DELIVERY"})
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		OTP string `json:"otp"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	var stored string
	h.pool.QueryRow(r.Context(), "SELECT delivery_otp FROM delivery_assignments WHERE id=$1", id).Scan(&stored)
	if stored != req.OTP {
		writeJSON(w, 400, map[string]string{"error": "invalid OTP"})
		return
	}
	var orderID string
	h.pool.QueryRow(r.Context(), "SELECT order_id FROM delivery_assignments WHERE id=$1", id).Scan(&orderID)
	_, _ = h.pool.Exec(r.Context(), "UPDATE delivery_assignments SET status='DELIVERED', delivered_at=now(), delivery_otp_verified=TRUE WHERE id=$1", id)
	_, _ = h.pool.Exec(r.Context(), "UPDATE orders SET status='DELIVERED', delivered_at=now(), updated_at=now(), payment_status=CASE WHEN payment_method='COD' THEN 'COD_COLLECTED' ELSE payment_status END WHERE id=$1", orderID)
	_, _ = h.pool.Exec(r.Context(), "INSERT INTO order_status_history(order_id,to_status,note) VALUES($1,'DELIVERED','delivery otp verified')", orderID)
	writeJSON(w, 200, map[string]string{"status": "DELIVERED"})
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	partnerID := r.PathValue("partner_id")
	var req struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_, _ = h.pool.Exec(r.Context(), "UPDATE delivery_partners SET current_location=ST_SetSRID(ST_MakePoint($1,$2),4326)::geography, last_location_at=now() WHERE id=$3", req.Lng, req.Lat, partnerID)
	writeJSON(w, 200, map[string]string{"status": "ok"})
}