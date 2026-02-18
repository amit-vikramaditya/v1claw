package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/events"
	"github.com/amit-vikramaditya/v1claw/pkg/logger"
	devsync "github.com/amit-vikramaditya/v1claw/pkg/sync"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			return true // Non-browser clients usually don't send Origin.
		}
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" {
			return false
		}
		return strings.EqualFold(u.Host, r.Host)
	},
}

type wsClient struct {
	id        string
	deviceID  string // Associated device ID for capability routing
	conn      *websocket.Conn
	send      chan []byte
	mu        sync.Mutex
	closeOnce sync.Once
	capReqs   map[string]chan CapabilityResponse // Pending capability requests
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ErrorC("api", fmt.Sprintf("WebSocket upgrade failed: %v", err))
		return
	}

	clientID := fmt.Sprintf("ws_%d", time.Now().UnixNano())
	client := &wsClient{
		id:      clientID,
		conn:    conn,
		send:    make(chan []byte, 64),
		capReqs: make(map[string]chan CapabilityResponse),
	}

	s.mu.Lock()
	s.wsClients[clientID] = client
	s.mu.Unlock()

	logger.InfoC("api", fmt.Sprintf("WebSocket client connected: %s", clientID))

	// Send welcome message.
	welcome := WSMessage{
		Type:      "connected",
		Data:      map[string]string{"client_id": clientID},
		Timestamp: time.Now(),
	}
	if data, err := json.Marshal(welcome); err == nil {
		client.send <- data
	}

	go s.wsWritePump(client)
	go s.wsReadPump(client)
}

func (s *Server) wsWritePump(client *wsClient) {
	defer func() {
		client.conn.Close()
		s.removeWSClient(client.id)
	}()

	for msg := range client.send {
		client.mu.Lock()
		err := client.conn.WriteMessage(websocket.TextMessage, msg)
		client.mu.Unlock()
		if err != nil {
			logger.DebugC("api", fmt.Sprintf("WebSocket write error: %v", err))
			return
		}
	}
}

func (s *Server) wsReadPump(client *wsClient) {
	defer func() {
		s.removeWSClient(client.id)
		client.conn.Close()
	}()

	client.conn.SetReadDeadline(time.Now().Add(300 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(300 * time.Second))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.DebugC("api", fmt.Sprintf("WebSocket read error: %v", err))
			}
			return
		}

		// Parse incoming message.
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case "chat":
			s.handleWSChat(client, wsMsg)
		case "ping":
			s.sendToClient(client, WSMessage{Type: "pong", Timestamp: time.Now()})
		case "capability_response":
			s.handleWSCapabilityResponse(client, wsMsg)
		}
	}
}

func (s *Server) handleWSChat(client *wsClient, msg WSMessage) {
	s.mu.RLock()
	handler := s.chatHandler
	s.mu.RUnlock()

	if handler == nil {
		s.sendToClient(client, WSMessage{
			Type:      "error",
			Data:      "chat handler not configured",
			Timestamp: time.Now(),
		})
		return
	}

	// Extract message text from data.
	text, _ := extractString(msg.Data, "message")
	sessionKey, _ := extractString(msg.Data, "session_key")
	if sessionKey == "" {
		sessionKey = "ws:" + client.id
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		response, err := handler(ctx, text, sessionKey)
		if err != nil {
			s.sendToClient(client, WSMessage{
				Type:      "error",
				Data:      err.Error(),
				Timestamp: time.Now(),
			})
			return
		}

		s.sendToClient(client, WSMessage{
			Type: "chat_response",
			Data: ChatResponse{
				Response:   response,
				SessionKey: sessionKey,
				Timestamp:  time.Now(),
			},
			Timestamp: time.Now(),
		})
	}()
}

func (s *Server) sendToClient(client *wsClient, msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case client.send <- data:
	default:
		// Client too slow, drop message.
	}
}

func (s *Server) broadcastEvent(event events.Event) {
	msg := WSMessage{
		Type:      "event",
		Data:      event,
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.wsClients {
		select {
		case client.send <- data:
		default:
		}
	}
}

func (s *Server) closeAllWS() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, client := range s.wsClients {
		client.conn.Close()
		client.closeOnce.Do(func() {
			close(client.send)
		})
		delete(s.wsClients, id)
	}
}

func extractString(data interface{}, key string) (string, bool) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return "", false
	}
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// handleWSCapabilityResponse processes capability responses from client devices.
func (s *Server) handleWSCapabilityResponse(client *wsClient, msg WSMessage) {
	reqID, _ := extractString(msg.Data, "request_id")
	if reqID == "" {
		return
	}

	var resp CapabilityResponse
	data, err := json.Marshal(msg.Data)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}

	client.mu.Lock()
	ch, ok := client.capReqs[reqID]
	if ok {
		delete(client.capReqs, reqID)
	}
	client.mu.Unlock()

	if ok {
		ch <- resp
	}
}

// RequestCapability sends a capability request to a specific device and waits for the response.
func (s *Server) RequestCapability(ctx context.Context, deviceID string, req CapabilityRequest) (*CapabilityResponse, error) {
	// Find the WS client associated with this device.
	s.mu.RLock()
	var target *wsClient
	for _, c := range s.wsClients {
		if c.deviceID == deviceID {
			target = c
			break
		}
	}
	s.mu.RUnlock()

	if target == nil {
		return nil, fmt.Errorf("device %s not connected via WebSocket", deviceID)
	}

	// Create response channel.
	respCh := make(chan CapabilityResponse, 1)
	target.mu.Lock()
	target.capReqs[req.RequestID] = respCh
	target.mu.Unlock()

	// Send request to client.
	s.sendToClient(target, WSMessage{
		Type:      "capability_request",
		Data:      req,
		Timestamp: time.Now(),
	})

	// Wait for response or timeout.
	select {
	case resp := <-respCh:
		return &resp, nil
	case <-ctx.Done():
		target.mu.Lock()
		delete(target.capReqs, req.RequestID)
		target.mu.Unlock()
		return nil, ctx.Err()
	}
}

// FindDeviceForCapability returns the best connected device for a capability.
func (s *Server) FindDeviceForCapability(cap string) (string, bool) {
	if s.registry == nil {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := s.registry.WithCapability(devsync.DeviceCapability(cap))
	for _, d := range devices {
		// Check if this device is connected via WS (not just registered).
		for _, c := range s.wsClients {
			if c.deviceID == d.ID {
				return d.ID, true
			}
		}
	}
	return "", false
}

// removeWSClient also unregisters the device and marks it offline.
func (s *Server) removeWSClient(id string) {
	s.mu.Lock()
	client, ok := s.wsClients[id]
	if ok {
		delete(s.wsClients, id)
	}
	s.mu.Unlock()

	if ok {
		// Mark device offline if associated.
		if client.deviceID != "" && s.registry != nil {
			s.registry.MarkOffline(client.deviceID)
			logger.InfoCF("api", "Device disconnected", map[string]interface{}{
				"device_id": client.deviceID, "ws_client": id,
			})
		}
		client.closeOnce.Do(func() {
			close(client.send)
		})
		logger.DebugC("api", fmt.Sprintf("WebSocket client disconnected: %s", id))
	}
}
