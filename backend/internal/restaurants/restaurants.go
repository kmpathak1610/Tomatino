package restaurants

import (
	"encoding/json"
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

// List restaurants
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, name, slug, description, cuisine_types, status, avg_rating, total_ratings
		FROM restaurants WHERE status='APPROVED' ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, name, slug string
		var desc *string
		var cuisines []string
		var status string
		var rating float64
		var total int
		rows.Scan(&id, &name, &slug, &desc, &cuisines, &status, &rating, &total)
		out = append(out, map[string]any{
			"id": id, "name": name, "slug": slug, "description": desc,
			"cuisine_types": cuisines, "status": status, "avg_rating": rating, "total_ratings": total,
		})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

// Create restaurant (onboarding)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Cuisines    []string `json:"cuisine_types"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	id := uuid.New().String()
	slug := id[:8]
	ownerID := r.Header.Get("X-User-ID")
	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO restaurants(id,owner_user_id,name,slug,description,cuisine_types,status)
		VALUES($1,$2,$3,$4,$5,$6,'PENDING')`,
		id, ownerID, req.Name, slug, req.Description, req.Cuisines)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "PENDING"})
}

// Approve (admin)
func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := h.pool.Exec(r.Context(), "UPDATE restaurants SET status='APPROVED', updated_at=$2 WHERE id=$1", id, time.Now())
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "APPROVED"})
}

// Get restaurant details with branches
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var name, slug, status string
	var desc *string
	var cuisines []string
	var rating float64
	var total int
	err := h.pool.QueryRow(r.Context(), `
		SELECT name, slug, description, cuisine_types, status, avg_rating, total_ratings
		FROM restaurants WHERE id=$1`, id).Scan(&name, &slug, &desc, &cuisines, &status, &rating, &total)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, map[string]any{
		"id": id, "name": name, "slug": slug, "description": desc,
		"cuisine_types": cuisines, "status": status, "avg_rating": rating, "total_ratings": total,
	})
}