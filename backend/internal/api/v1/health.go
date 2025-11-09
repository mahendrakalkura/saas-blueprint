package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/database"
)

var startTime = time.Now()

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Services  map[string]ServiceInfo `json:"services"`
	System    SystemInfo             `json:"system"`
}

type ServiceInfo struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemoryUsageMB float64 `json:"memory_usage_mb"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]ServiceInfo)

	// Check database
	dbInfo := ServiceInfo{Status: "healthy", Details: make(map[string]string)}
	if err := h.db.HealthCheck(r.Context()); err != nil {
		dbInfo.Status = "unhealthy"
		dbInfo.Details["error"] = err.Error()
	} else {
		// Get database stats
		stats := h.db.DB.Stats()
		dbInfo.Details["open_connections"] = string(rune(stats.OpenConnections))
		dbInfo.Details["in_use"] = string(rune(stats.InUse))
		dbInfo.Details["idle"] = string(rune(stats.Idle))
	}
	services["database"] = dbInfo

	// Determine overall status
	status := "ok"
	for _, serviceInfo := range services {
		if serviceInfo.Status == "unhealthy" {
			status = "degraded"
			break
		}
	}

	// Get system info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemInfo := SystemInfo{
		GoVersion:     runtime.Version(),
		NumGoroutine:  runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		MemoryUsageMB: float64(m.Alloc) / 1024 / 1024,
	}

	uptime := time.Since(startTime)

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0", // Should come from config or build info
		Uptime:    uptime.String(),
		Services:  services,
		System:    systemInfo,
	}

	statusCode := http.StatusOK
	if status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	if err := h.db.HealthCheck(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
