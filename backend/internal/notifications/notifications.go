package notifications

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	rows, _ := h.pool.Query(r.Context(), "SELECT id, title, body, type, read_at, created_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50", userID)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, title, body, typ string
		var readAt, created interface{}
		rows.Scan(&id, &title, &body, &typ, &readAt, &created)
		out = append(out, map[string]any{"id": id, "title": title, "body": body, "type": typ, "read_at": readAt, "created_at": created})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = h.pool.Exec(r.Context(), "UPDATE notifications SET read_at=now() WHERE id=$1", id)
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		Type   string `json:"type"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Type == "" {
		req.Type = "INFO"
	}
	id := uuid.New().String()
	_, err := h.pool.Exec(r.Context(), "INSERT INTO notifications(id,user_id,title,body,type) VALUES($1,$2,$3,$4,$5)", id, req.UserID, req.Title, req.Body, req.Type)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}