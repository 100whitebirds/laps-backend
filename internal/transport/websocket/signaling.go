package websocket

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"laps/internal/domain"
)

// SignalingMessage represents a WebRTC signaling message
type SignalingMessage struct {
	Type      string      `json:"type"`
	SessionID string      `json:"session_id"`
	From      int64       `json:"from"`
	To        int64       `json:"to"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID     int64
	UserID int64
	Role   domain.UserRole
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *SignalingHub
}

// SignalingHub maintains the set of active clients and broadcasts messages
type SignalingHub struct {
	// Registered clients by user ID
	clients map[int64]*Client

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Active call sessions by session ID
	sessions map[string]*CallSession

	// Logger
	logger *zap.Logger

	// Mutex for thread safety
	mutex sync.RWMutex
}

// CallSession represents an active call session
type CallSession struct {
	ID           string    `json:"id"`
	ClientID     int64     `json:"client_id"`
	SpecialistID int64     `json:"specialist_id"`
	AppointmentID *int64   `json:"appointment_id,omitempty"`
	Status       string    `json:"status"` // waiting, active, ended
	CreatedAt    time.Time `json:"created_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should validate the origin
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// NewSignalingHub creates a new signaling hub
func NewSignalingHub(logger *zap.Logger) *SignalingHub {
	return &SignalingHub{
		clients:    make(map[int64]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		sessions:   make(map[string]*CallSession),
		logger:     logger,
	}
}

// Run starts the signaling hub
func (h *SignalingHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.UserID] = client
			h.mutex.Unlock()
			h.logger.Info("Client connected", 
				zap.Int64("user_id", client.UserID), 
				zap.String("role", string(client.Role)))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mutex.Unlock()
			h.logger.Info("Client disconnected", zap.Int64("user_id", client.UserID))

		case message := <-h.broadcast:
			var msg SignalingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				h.logger.Error("Failed to unmarshal message", zap.Error(err))
				continue
			}

			h.handleSignalingMessage(&msg)
		}
	}
}

// handleSignalingMessage processes incoming signaling messages
func (h *SignalingHub) handleSignalingMessage(msg *SignalingMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	switch msg.Type {
	case "call-offer":
		h.handleCallOffer(msg)
	case "call-answer":
		h.handleCallAnswer(msg)
	case "ice-candidate":
		h.handleIceCandidate(msg)
	case "call-end":
		h.handleCallEnd(msg)
	case "ping":
		h.handlePing(msg)
	default:
		h.logger.Warn("Unknown message type", zap.String("type", msg.Type))
	}
}

// handleCallOffer processes call offer messages
func (h *SignalingHub) handleCallOffer(msg *SignalingMessage) {
	// Create new call session
	session := &CallSession{
		ID:           msg.SessionID,
		ClientID:     msg.From,
		SpecialistID: msg.To,
		Status:       "waiting",
		CreatedAt:    time.Now(),
	}

	h.sessions[msg.SessionID] = session

	// Forward offer to target user
	if targetClient, exists := h.clients[msg.To]; exists {
		h.sendMessageToClient(targetClient, msg)
		h.logger.Info("Call offer forwarded", 
			zap.String("session_id", msg.SessionID),
			zap.Int64("from", msg.From),
			zap.Int64("to", msg.To))
	} else {
		h.logger.Warn("Target user not connected", zap.Int64("user_id", msg.To))
		// Send error back to caller
		errorMsg := &SignalingMessage{
			Type:      "call-error",
			SessionID: msg.SessionID,
			From:      msg.To,
			To:        msg.From,
			Data:      map[string]string{"error": "User not available"},
			Timestamp: time.Now(),
		}
		if callerClient, exists := h.clients[msg.From]; exists {
			h.sendMessageToClient(callerClient, errorMsg)
		}
	}
}

// handleCallAnswer processes call answer messages
func (h *SignalingHub) handleCallAnswer(msg *SignalingMessage) {
	// Update session status
	if session, exists := h.sessions[msg.SessionID]; exists {
		session.Status = "active"
	}

	// Forward answer to caller
	if callerClient, exists := h.clients[msg.To]; exists {
		h.sendMessageToClient(callerClient, msg)
		h.logger.Info("Call answer forwarded", 
			zap.String("session_id", msg.SessionID),
			zap.Int64("from", msg.From),
			zap.Int64("to", msg.To))
	}
}

// handleIceCandidate processes ICE candidate messages
func (h *SignalingHub) handleIceCandidate(msg *SignalingMessage) {
	// Forward ICE candidate to the other peer
	if targetClient, exists := h.clients[msg.To]; exists {
		h.sendMessageToClient(targetClient, msg)
	}
}

// handleCallEnd processes call end messages
func (h *SignalingHub) handleCallEnd(msg *SignalingMessage) {
	// Update session status
	if session, exists := h.sessions[msg.SessionID]; exists {
		session.Status = "ended"
		now := time.Now()
		session.EndedAt = &now
	}

	// Forward end message to the other peer
	if targetClient, exists := h.clients[msg.To]; exists {
		h.sendMessageToClient(targetClient, msg)
	}

	h.logger.Info("Call ended", zap.String("session_id", msg.SessionID))
}

// handlePing processes ping messages for connection keepalive
func (h *SignalingHub) handlePing(msg *SignalingMessage) {
	pongMsg := &SignalingMessage{
		Type:      "pong",
		SessionID: msg.SessionID,
		From:      msg.To,
		To:        msg.From,
		Timestamp: time.Now(),
	}

	if client, exists := h.clients[msg.From]; exists {
		h.sendMessageToClient(client, pongMsg)
	}
}

// sendMessageToClient sends a message to a specific client
func (h *SignalingHub) sendMessageToClient(client *Client, msg *SignalingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	select {
	case client.Send <- data:
	default:
		delete(h.clients, client.UserID)
		close(client.Send)
	}
}

// HandleWebSocket handles WebSocket connections
func (h *SignalingHub) HandleWebSocket(c *gin.Context) {
	// Get user ID and role from JWT token (passed as query parameter for WebSocket)
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	// TODO: Validate JWT token and extract user info
	// For now, we'll get user_id from query parameter (in production, extract from JWT)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	role := domain.UserRole(c.DefaultQuery("role", "client"))

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Create client
	client := &Client{
		UserID: userID,
		Role:   role,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h,
	}

	// Register client
	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		// Parse and validate message
		var msg SignalingMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.Hub.logger.Error("Failed to unmarshal message", zap.Error(err))
			continue
		}

		// Set sender info
		msg.From = c.UserID
		msg.Timestamp = time.Now()

		// Re-marshal with corrected info
		correctedMessage, err := json.Marshal(msg)
		if err != nil {
			c.Hub.logger.Error("Failed to marshal corrected message", zap.Error(err))
			continue
		}

		c.Hub.broadcast <- correctedMessage
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetActiveSessions returns all active call sessions
func (h *SignalingHub) GetActiveSessions() map[string]*CallSession {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	sessions := make(map[string]*CallSession)
	for id, session := range h.sessions {
		if session.Status == "active" || session.Status == "waiting" {
			sessions[id] = session
		}
	}
	return sessions
}

// IsUserConnected checks if a user is currently connected
func (h *SignalingHub) IsUserConnected(userID int64) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	_, exists := h.clients[userID]
	return exists
} 