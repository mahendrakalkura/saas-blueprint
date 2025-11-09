package websocket

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should validate the origin
		// For now, allow all origins
		return true
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeWS handles WebSocket requests from clients
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	// Create client
	client := &Client{
		ID:     uuid.New().String(),
		UserID: userCtx.UserID,
		Hub:    h.hub,
		Conn:   NewConn(ws),
		Send:   make(chan []byte, 256),
	}

	// Register client with hub
	client.Hub.register <- client

	// Start read and write pumps in goroutines
	go client.WritePump()
	go client.ReadPump()
}

// GetStats returns current WebSocket statistics
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_connections": h.hub.GetTotalConnections(),
		"unique_users":      h.hub.GetConnectedUserCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON encoding
	response := `{"total_connections":` + string(rune(stats["total_connections"].(int))) + `,"unique_users":` + string(rune(stats["unique_users"].(int))) + `}`
	w.Write([]byte(response))
}
