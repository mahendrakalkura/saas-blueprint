package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/database"
)

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]string)

	// Check database
	if err := h.db.HealthCheck(r.Context()); err != nil {
		services["database"] = "unhealthy"
	} else {
		services["database"] = "healthy"
	}

	status := "ok"
	for _, serviceStatus := range services {
		if serviceStatus == "unhealthy" {
			status = "degraded"
			break
		}
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  services,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
