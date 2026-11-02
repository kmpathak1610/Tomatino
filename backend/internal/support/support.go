package support

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

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		Subject  string  `json:"subject"`
		OrderID  *string `json:"order_id"`
		Priority string  `json:"priority"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Priority == "" {
		req.Priority = "MEDIUM"
	}
	id := uuid.New().String()
	_, err := h.pool.Exec(r.Context(), "INSERT INTO support_tickets(id,user_id,order_id,subject,priority) VALUES($1,$2,$3,$4,$5)", id, userID, req.OrderID, req.Subject, req.Priority)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

func (h *Handler) AddMessage(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	ticketID := r.PathValue("id")
	var req struct {
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	id := uuid.New().String()
	_, err := h.pool.Exec(r.Context(), "INSERT INTO support_messages(id,ticket_id,sender_id,message) VALUES($1,$2,$3,$4)", id, ticketID, userID, req.Message)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

func (h *Handler) ListMyTickets(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	rows, _ := h.pool.Query(r.Context(), "SELECT id, subject, status, priority, created_at FROM support_tickets WHERE user_id=$1 ORDER BY created_at DESC", userID)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, subject, status, priority string
		var created interface{}
		rows.Scan(&id, &subject, &status, &priority, &created)
		out = append(out, map[string]any{"id": id, "subject": subject, "status": status, "priority": priority, "created_at": created})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}