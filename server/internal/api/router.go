package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/moabdelazem/k8s-app/internal/api/handlers"
	"github.com/moabdelazem/k8s-app/pkg/logger"
	"go.uber.org/zap"
)

func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(LoggingMiddleware)

	r.Get("/health", handlers.Health)

	// Liveness and Readiness Probes For Kubernetes
	r.Get("/live", handlers.LivenessProbe)
	r.Get("/ready", handlers.ReadinessProbe)

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
