package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fooddelivery/internal/admin"
	"fooddelivery/internal/auth"
	"fooddelivery/internal/cart"
	"fooddelivery/internal/coupons"
	"fooddelivery/internal/devicetokens"
	"fooddelivery/internal/db"
	"fooddelivery/internal/delivery"
	"fooddelivery/internal/favourites"
	"fooddelivery/internal/menu"
	"fooddelivery/internal/notifications"
	"fooddelivery/internal/orders"
	"fooddelivery/internal/partner"
	"fooddelivery/internal/payments"
	"fooddelivery/internal/restaurants"
	"fooddelivery/internal/reviews"
	"fooddelivery/internal/support"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pool, err := db.NewPool(context.Background())
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	authSvc := auth.NewService(pool)
	cartH := cart.NewHandler(pool)
	couponH := coupons.NewHandler(pool)
	deviceH := devicetokens.NewHandler(pool)
	favH := favourites.NewHandler(pool)
	restH := restaurants.NewHandler(pool)
	reviewH := reviews.NewHandler(pool)
	supportH := support.NewHandler(pool)
	menuH := menu.NewHandler(pool)
	notifH := notifications.NewHandler(pool)
	orderH := orders.NewHandler(pool)
	partnerH := partner.NewHandler(pool)
	payH := payments.NewHandler(pool)
	delH := delivery.NewHandler(pool)
	adminH := admin.NewHandler(pool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// auth
		r.Post("/auth/otp/send", authSvc.SendOTP)
		r.Post("/auth/otp/verify", authSvc.VerifyOTP)
		r.Post("/auth/staff/login", authSvc.StaffLogin)

		// restaurants (public read)
		r.Get("/restaurants", restH.List)
		r.Get("/restaurants/{id}", restH.Get)
		r.With(authSvc.Middleware).Post("/restaurants", restH.Create)
		r.With(authSvc.Middleware).Post("/restaurants/{id}/approve", restH.Approve)

		r.With(authSvc.Middleware).Post("/reviews", reviewH.Create)

		r.Route("/device-tokens", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Post("/", deviceH.Register)
		})

		r.Route("/favourites", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Get("/", favH.List)
			r.Post("/", favH.Add)
			r.Delete("/{restaurant_id}", favH.Remove)
		})

		r.Route("/notifications", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Get("/", notifH.List)
			r.Post("/{id}/read", notifH.MarkRead)
		})
		r.Post("/admin/notifications", notifH.Create)

		r.Route("/support", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Get("/tickets", supportH.ListMyTickets)
			r.Post("/tickets", supportH.CreateTicket)
			r.Post("/tickets/{id}/messages", supportH.AddMessage)
		})

		// cart
		r.Route("/cart", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Get("/", cartH.Get)
			r.Post("/items", cartH.AddItem)
			r.Delete("/items/{id}", cartH.RemoveItem)
			r.Delete("/", cartH.Clear)
		})

		// coupons
		r.Route("/coupons", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Post("/validate", couponH.Validate)
			r.Post("/", couponH.Create)
		})

		// menu
		r.Get("/reviews", reviewH.ListByRestaurant)

		r.Get("/menu", menuH.ListByRestaurant)
		r.With(authSvc.Middleware).Post("/menu/items", menuH.CreateItem)

		// orders
		r.Route("/orders", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Post("/", orderH.Create)
			r.Get("/", orderH.ListByUser)
			r.Get("/{id}", orderH.Get)
			r.Patch("/{id}/status", orderH.UpdateStatus)
		})

		// payments
		r.Route("/payments", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Post("/create-order", payH.CreateOrder)
			r.Post("/verify", payH.Verify)
		})
		r.Post("/payments/webhook", payH.Webhook)

		// delivery
		r.Route("/delivery", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Get("/assignments", delH.MyAssignments)
			r.Get("/available", delH.Available)
			r.Post("/claim/{order_id}", delH.Claim)
			r.Post("/register", delH.Register)
			r.Post("/partners/{id}/approve", delH.Approve)
			r.Post("/assign", delH.Assign)
			r.Post("/assignments/{id}/accept", delH.Accept)
			r.Post("/assignments/{id}/pickup", delH.Pickup)
			r.Post("/assignments/{id}/out-for-delivery", delH.OutForDelivery)
			r.Post("/assignments/{id}/complete", delH.Complete)
			r.Post("/partners/{id}/location", delH.UpdateLocation)
		})

		// restaurant partner
		r.Route("/partner", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Get("/restaurant", partnerH.Restaurant)
			r.Get("/orders", partnerH.Orders)
			r.Get("/analytics", partnerH.Analytics)
			r.Get("/menu", partnerH.Menu)
			r.Patch("/menu/{id}/availability", partnerH.ToggleAvailability)
		})

		// admin
		r.Route("/admin", func(r chi.Router) {
			r.Use(authSvc.Middleware)
			r.Get("/stats", adminH.Stats)
			r.Get("/orders", adminH.ListOrders)
			r.Get("/users", adminH.ListUsers)
			r.Get("/restaurants", adminH.ListRestaurants)
		})
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		logger.Info("starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	fmt.Println("server stopped")
}