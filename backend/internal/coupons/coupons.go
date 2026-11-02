package coupons

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

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		Code            string  `json:"code"`
		OrderAmount     float64 `json:"order_amount"`
		RestaurantID    string  `json:"restaurant_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var id, dtype string
	var dval, minOrder, maxDisc float64
	var usageLimit, usedCount, perUser int
	var restID *string
	var validFrom, validUntil time.Time
	var active bool
	err := h.pool.QueryRow(r.Context(), `
		SELECT id, discount_type, discount_value, min_order_amount, COALESCE(max_discount,0),
		       usage_limit, used_count, per_user_limit, restaurant_id, valid_from, valid_until, is_active
		FROM coupons WHERE code=$1`, req.Code).Scan(
		&id, &dtype, &dval, &minOrder, &maxDisc, &usageLimit, &usedCount, &perUser, &restID, &validFrom, &validUntil, &active)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "coupon not found"})
		return
	}
	now := time.Now()
	if !active || now.Before(validFrom) || now.After(validUntil) {
		writeJSON(w, 400, map[string]string{"error": "coupon not valid"})
		return
	}
	if req.OrderAmount < minOrder {
		writeJSON(w, 400, map[string]string{"error": "minimum order not met"})
		return
	}
	if restID != nil && *restID != req.RestaurantID {
		writeJSON(w, 400, map[string]string{"error": "coupon not applicable to this restaurant"})
		return
	}
	if usageLimit > 0 && usedCount >= usageLimit {
		writeJSON(w, 400, map[string]string{"error": "coupon usage limit reached"})
		return
	}
	// per-user
	var userUses int
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id=$1 AND user_id=$2", id, userID).Scan(&userUses)
	if userUses >= perUser {
		writeJSON(w, 400, map[string]string{"error": "per-user limit reached"})
		return
	}
	var discount float64
	if dtype == "PERCENTAGE" {
		discount = req.OrderAmount * dval / 100
		if maxDisc > 0 && discount > maxDisc {
			discount = maxDisc
		}
	} else {
		discount = dval
		if discount > req.OrderAmount {
			discount = req.OrderAmount
		}
	}
	writeJSON(w, 200, map[string]any{"coupon_id": id, "discount": discount, "code": req.Code})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code           string  `json:"code"`
		Description    string  `json:"description"`
		DiscountType   string  `json:"discount_type"`
		DiscountValue  float64 `json:"discount_value"`
		MinOrderAmount float64 `json:"min_order_amount"`
		MaxDiscount    float64 `json:"max_discount"`
		UsageLimit     int     `json:"usage_limit"`
		PerUserLimit   int     `json:"per_user_limit"`
		RestaurantID   *string `json:"restaurant_id"`
		ValidFrom      string  `json:"valid_from"`
		ValidUntil     string  `json:"valid_until"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.PerUserLimit == 0 {
		req.PerUserLimit = 1
	}
	vf, _ := time.Parse(time.RFC3339, req.ValidFrom)
	vu, _ := time.Parse(time.RFC3339, req.ValidUntil)
	id := uuid.New().String()
	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO coupons(id,code,description,discount_type,discount_value,min_order_amount,max_discount,usage_limit,per_user_limit,restaurant_id,valid_from,valid_until,is_active)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,TRUE)`,
		id, req.Code, req.Description, req.DiscountType, req.DiscountValue, req.MinOrderAmount, req.MaxDiscount, req.UsageLimit, req.PerUserLimit, req.RestaurantID, vf, vu)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}