package payments

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

// Create payment order (mock razorpay)
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID string  `json:"order_id"`
		Amount  float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	providerOrderID := "order_mock_" + req.OrderID[:8]
	_, _ = h.pool.Exec(r.Context(), `
		INSERT INTO payments(order_id,provider,provider_order_id,amount,status)
		VALUES($1,'RAZORPAY',$2,$3,'PENDING')`, req.OrderID, providerOrderID, req.Amount)
	writeJSON(w, 200, map[string]any{
		"provider": "razorpay", "provider_order_id": providerOrderID,
		"amount": req.Amount, "currency": "INR", "key_id": "rzp_test_mock",
	})
}

// Verify payment (mock)
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID          string `json:"order_id"`
		ProviderPaymentID string `json:"provider_payment_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_, _ = h.pool.Exec(r.Context(), "UPDATE payments SET provider_payment_id=$1,status='PAID' WHERE order_id=$2", req.ProviderPaymentID, req.OrderID)
	_, _ = h.pool.Exec(r.Context(), "UPDATE orders SET payment_status='PAID' WHERE id=$1", req.OrderID)
	writeJSON(w, 200, map[string]string{"status": "PAID"})
}

// Webhook (idempotent)
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	json.NewDecoder(r.Body).Decode(&payload)
	eventID, _ := payload["event_id"].(string)
	if eventID == "" {
		eventID = "evt_" + reqID()
	}
	// idempotency
	var exists bool
	h.pool.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM payment_webhooks WHERE event_id=$1)", eventID).Scan(&exists)
	if exists {
		writeJSON(w, 200, map[string]string{"status": "duplicate"})
		return
	}
	b, _ := json.Marshal(payload)
	_, _ = h.pool.Exec(r.Context(), "INSERT INTO payment_webhooks(event_id,payload,processed) VALUES($1,$2,TRUE)", eventID, string(b))
	writeJSON(w, 200, map[string]string{"status": "processed"})
}

func reqID() string { return "x" }