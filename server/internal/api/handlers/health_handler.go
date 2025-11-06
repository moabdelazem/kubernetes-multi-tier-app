package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/moabdelazem/k8s-app/internal/database"
	"github.com/moabdelazem/k8s-app/pkg/logger"
	"github.com/moabdelazem/k8s-app/pkg/response"
	"go.uber.org/zap"
)

var startTime = time.Now()

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version"`
	System    SystemInfo        `json:"system"`
	Database  *DatabaseInfo     `json:"database,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// SystemInfo contains system information
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
}

// DatabaseInfo contains database connection pool information
type DatabaseInfo struct {
	Status            string `json:"status"`
	OpenConnections   int    `json:"open_connections"`
	InUse             int    `json:"in_use"`
	Idle              int    `json:"idle"`
	WaitCount         int64  `json:"wait_count"`
	WaitDuration      string `json:"wait_duration"`
	MaxIdleClosed     int64  `json:"max_idle_closed"`
	MaxLifetimeClosed int64  `json:"max_lifetime_closed"`
}

// Health handles the health check endpoint
func Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)

	healthData := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    uptime.String(),
		Version:   "1.0.0",
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
	}

	// Add database health information
	if err := database.Ping(); err != nil {
		healthData.Database = &DatabaseInfo{
			Status: "unhealthy",
		}
		logger.Warn("Database health check failed in /health endpoint",
			zap.Error(err),
		)
	} else {
		stats := database.Stats()
		healthData.Database = &DatabaseInfo{
			Status:            "healthy",
			OpenConnections:   stats.OpenConnections,
			InUse:             stats.InUse,
			Idle:              stats.Idle,
			WaitCount:         stats.WaitCount,
			WaitDuration:      stats.WaitDuration.String(),
			MaxIdleClosed:     stats.MaxIdleClosed,
			MaxLifetimeClosed: stats.MaxLifetimeClosed,
		}
	}

	response.Success(w, "", healthData)
}

// LivenessProbe is a simple liveness check for Kubernetes
// Returns 200 if the application is running
func LivenessProbe(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "alive",
	})
}

// ReadinessProbe checks if the application is ready to serve traffic
func ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	isReady := true

	// Database health check
	if err := database.Ping(); err != nil {
		checks["database"] = "unhealthy"
		isReady = false
		logger.Error("Database health check failed",
			zap.Error(err),
		)
	} else {
		checks["database"] = "healthy"

		// Add connection pool stats
		stats := database.Stats()
		logger.Debug("Database connection pool stats",
			zap.Int("open_connections", stats.OpenConnections),
			zap.Int("in_use", stats.InUse),
			zap.Int("idle", stats.Idle),
		)
	}

	checks["api"] = "ready"

	if !isReady {
		response.JSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "not ready",
			"checks": checks,
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "ready",
		"checks": checks,
	})
}
