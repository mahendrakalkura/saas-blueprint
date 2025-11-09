package websocket

import (
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"
)

// Message represents a WebSocket message
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID     string
	UserID string
	Hub    *Hub
	Conn   *Conn
	Send   chan []byte
}

// Hub manages all WebSocket connections and broadcasts
type Hub struct {
	// Registered clients
	clients map[string]*Client

	// Clients by user ID for targeted messages
	userClients map[string]map[string]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients
	broadcast chan []byte

	// Send message to specific user
	userMessage chan *UserMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// UserMessage represents a message targeted to a specific user
type UserMessage struct {
	UserID  string
	Message []byte
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		userClients: make(map[string]map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte, 256),
		userMessage: make(chan *UserMessage, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case userMsg := <-h.userMessage:
			h.sendToUser(userMsg.UserID, userMsg.Message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	// Add to user's clients map
	if _, ok := h.userClients[client.UserID]; !ok {
		h.userClients[client.UserID] = make(map[string]*Client)
	}
	h.userClients[client.UserID][client.ID] = client

	log.Info().
		Str("client_id", client.ID).
		Str("user_id", client.UserID).
		Int("total_clients", len(h.clients)).
		Msg("Client connected")
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)

		// Remove from user's clients map
		if userClients, ok := h.userClients[client.UserID]; ok {
			delete(userClients, client.ID)
			if len(userClients) == 0 {
				delete(h.userClients, client.UserID)
			}
		}

		close(client.Send)

		log.Info().
			Str("client_id", client.ID).
			Str("user_id", client.UserID).
			Int("total_clients", len(h.clients)).
			Msg("Client disconnected")
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client.ID)
		}
	}
}

func (h *Hub) sendToUser(userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if userClients, ok := h.userClients[userID]; ok {
		for _, client := range userClients {
			select {
			case client.Send <- message:
			default:
				// Client's send channel is full, skip
				log.Warn().
					Str("client_id", client.ID).
					Str("user_id", userID).
					Msg("Client send buffer full")
			}
		}
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(msgType string, payload map[string]interface{}) error {
	message := Message{
		Type:    msgType,
		Payload: payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- data
	return nil
}

// SendToUser sends a message to all connections of a specific user
func (h *Hub) SendToUser(userID string, msgType string, payload map[string]interface{}) error {
	message := Message{
		Type:    msgType,
		Payload: payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.userMessage <- &UserMessage{
		UserID:  userID,
		Message: data,
	}

	return nil
}

// GetConnectedUserCount returns the number of unique users connected
func (h *Hub) GetConnectedUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userClients)
}

// GetTotalConnections returns the total number of client connections
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
