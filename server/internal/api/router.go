package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/moabdelazem/k8s-app/internal/api/handlers"
	"github.com/moabdelazem/k8s-app/internal/config"
	"github.com/moabdelazem/k8s-app/internal/repository"
	"github.com/moabdelazem/k8s-app/internal/service"
	"github.com/moabdelazem/k8s-app/pkg/logger"
	"go.uber.org/zap"
)

func SetupRoutes(db *sql.DB, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// CORS middleware - configured from environment variables
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.CORS.ExposedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	logger.Info("CORS configured",
		zap.Strings("allowed_origins", cfg.CORS.AllowedOrigins),
		zap.Bool("allow_credentials", cfg.CORS.AllowCredentials),
	)

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(LoggingMiddleware)

	// Health endpoints
	r.Get("/health", handlers.Health)
	r.Get("/live", handlers.LivenessProbe)
	r.Get("/ready", handlers.ReadinessProbe)

	// Initialize poll dependencies
	pollRepo := repository.NewPollRepository(db)
	pollService := service.NewPollService(pollRepo)
	pollHandler := handlers.NewPollHandler(pollService)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Poll routes
		r.Route("/polls", func(r chi.Router) {
			r.Post("/", pollHandler.CreatePoll)          // Create poll
			r.Get("/", pollHandler.ListPolls)            // List polls
			r.Get("/{id}", pollHandler.GetPoll)          // Get poll with results
			r.Post("/{id}/vote", pollHandler.VoteOnPoll) // Vote on poll
			r.Delete("/{id}", pollHandler.DeletePoll)    // Delete poll
		})
	})

	return r
}

// LoggingMiddleware logs incoming requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Incoming request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)
		next.ServeHTTP(w, r)
	})
}
