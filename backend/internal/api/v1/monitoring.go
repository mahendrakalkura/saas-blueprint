package v1

import (
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
)

type MonitoringHandler struct {
	inspector *asynq.Inspector
}

func NewMonitoringHandler(cfg *config.Config) *MonitoringHandler {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: cfg.Redis.Addr,
	})

	return &MonitoringHandler{
		inspector: inspector,
	}
}

// GetQueueStats returns statistics about all queues
func (h *MonitoringHandler) GetQueueStats(w http.ResponseWriter, r *http.Request) {
	queues, err := h.inspector.Queues()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get queue info")
		return
	}

	stats := make(map[string]interface{})
	for _, queue := range queues {
		queueStats, err := h.inspector.GetQueueInfo(queue)
		if err != nil {
			continue
		}

		stats[queue] = map[string]interface{}{
			"active":     queueStats.Active,
			"pending":    queueStats.Pending,
			"scheduled":  queueStats.Scheduled,
			"retry":      queueStats.Retry,
			"archived":   queueStats.Archived,
			"completed":  queueStats.Completed,
			"aggregating": queueStats.Aggregating,
			"size":       queueStats.Size,
			"latency":    queueStats.Latency.String(),
			"memory_usage": queueStats.MemoryUsage,
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"queues": stats,
	})
}

// GetServerInfo returns information about the worker server
func (h *MonitoringHandler) GetServerInfo(w http.ResponseWriter, r *http.Request) {
	servers, err := h.inspector.Servers()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get server info")
		return
	}

	serversInfo := make([]map[string]interface{}, 0)
	for _, server := range servers {
		serversInfo = append(serversInfo, map[string]interface{}{
			"host":              server.Host,
			"pid":               server.PID,
			"server_id":         server.ServerID,
			"concurrency":       server.Concurrency,
			"queues":            server.Queues,
			"strict_priority":   server.StrictPriority,
			"status":            server.Status,
			"started_at":        server.Started,
			"active_workers":    server.ActiveWorkers,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"servers": serversInfo,
	})
}

// GetScheduledTasks returns all scheduled tasks
func (h *MonitoringHandler) GetScheduledTasks(w http.ResponseWriter, r *http.Request) {
	schedulerEntries, err := h.inspector.SchedulerEntries()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get scheduled tasks")
		return
	}

	entries := make([]map[string]interface{}, 0)
	for _, entry := range schedulerEntries {
		entries = append(entries, map[string]interface{}{
			"id":       entry.ID,
			"spec":     entry.Spec,
			"task":     entry.Task.Type(),
			"opts":     entry.Opts,
			"next_run": entry.Next,
			"prev_run": entry.Prev,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"scheduled_tasks": entries,
	})
}

// Close closes the inspector
func (h *MonitoringHandler) Close() error {
	return h.inspector.Close()
}
