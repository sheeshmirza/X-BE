package router

import (
	"net/http"

	"X-BE/internal/config"
	"X-BE/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(cfg config.Config, handlers *handlers.Handlers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/health", handlers.Health)
	registerAPIRoutes(r, handlers)
	return cors(r, cfg)
}

func registerAPIRoutes(r chi.Router, handlers *handlers.Handlers) {
	r.Route("/api", func(api chi.Router) {
		api.Route("/career", func(career chi.Router) {
			career.Post("/create", handlers.CareerCreate)
		})
		api.Route("/newsletter", func(newsletter chi.Router) {
			newsletter.Post("/create", handlers.NewsletterCreate)
			newsletter.Delete("/delete", handlers.NewsletterDelete)
		})
		api.Route("/sale", func(sale chi.Router) {
			sale.Post("/create", handlers.SaleCreate)
		})
		registerZohoRoutes(api, handlers)
	})
}

func registerZohoRoutes(api chi.Router, handlers *handlers.Handlers) {
	api.Route("/zoho", func(zohoRoute chi.Router) {
		zohoRoute.Get("/auth", handlers.ZohoAuth)
		zohoRoute.Get("/oauth/callback", handlers.ZohoOAuthCallback)
	})
}

func cors(next http.Handler, cfg config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := cfg.AllowedOrigin
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
