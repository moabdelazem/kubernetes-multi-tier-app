package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/moabdelazem/k8s-app/pkg/response"
)

var startTime = time.Now()

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version"`
	System    SystemInfo        `json:"system"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// SystemInfo contains system information
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
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

	// TODO: Database check
	// if err := db.Ping(); err != nil {
	//     checks["database"] = "unhealthy"
	//     isReady = false
	// } else {
	//     checks["database"] = "healthy"
	// }

	// For now, just return ready
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
