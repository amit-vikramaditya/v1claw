package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/bus"
	"github.com/amit-vikramaditya/v1claw/pkg/events"
	"github.com/amit-vikramaditya/v1claw/pkg/logger"
	"github.com/amit-vikramaditya/v1claw/pkg/state"
)

// Server provides HTTP and WebSocket API endpoints for V1Claw.
type Server struct {
	mu         sync.RWMutex
	addr       string
	apiKey     string
	msgBus     *bus.MessageBus
	router     *events.Router
	stateMgr   *state.Manager
	httpServer *http.Server
	wsClients  map[string]*wsClient
	chatHandler ChatHandler
}

// ChatHandler processes chat messages from the API and returns responses.
type ChatHandler func(ctx context.Context, message, sessionKey string) (string, error)

// Config holds API server configuration.
type Config struct {
	// Addr is the listen address (default ":18791").
	Addr string `json:"addr"`
	// APIKey is the authentication key for API access. Empty disables auth.
	APIKey string `json:"api_key"`
}

// NewServer creates a new API server.
func NewServer(cfg Config, msgBus *bus.MessageBus, router *events.Router, stateMgr *state.Manager) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":18791"
	}
	return &Server{
		addr:      cfg.Addr,
		apiKey:    cfg.APIKey,
		msgBus:    msgBus,
		router:    router,
		stateMgr:  stateMgr,
		wsClients: make(map[string]*wsClient),
	}
}

// SetChatHandler sets the function that processes chat messages.
func (s *Server) SetChatHandler(handler ChatHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatHandler = handler
}

// Start begins serving HTTP requests.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/chat", s.authMiddleware(s.handleChat))
	mux.HandleFunc("/api/v1/status", s.authMiddleware(s.handleStatus))
	mux.HandleFunc("/api/v1/users", s.authMiddleware(s.handleUsers))
	mux.HandleFunc("/api/v1/events", s.authMiddleware(s.handleEvents))
	mux.HandleFunc("/api/v1/ws", s.authMiddleware(s.handleWebSocket))

	// Health endpoints (no auth)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Subscribe to events for WebSocket broadcast.
	if s.router != nil {
		s.router.Subscribe("api_ws_broadcast", func(ctx context.Context, event events.Event) error {
			s.broadcastEvent(event)
			return nil
		})
	}

	logger.InfoC("api", fmt.Sprintf("API server starting on %s", s.addr))

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(shutdownCtx)
	}()

	err := s.httpServer.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}
	s.closeAllWS()
	logger.InfoC("api", "API server stopped")
}

// --- Middleware ---

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.apiKey == "" {
			next(w, r)
			return
		}

		// Check Authorization header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			// Also check query param for WebSocket connections
			auth = r.URL.Query().Get("api_key")
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token != s.apiKey {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid or missing API key"})
			return
		}

		next(w, r)
	}
}

// --- Handlers ---

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "message is required"})
		return
	}

	if req.SessionKey == "" {
		req.SessionKey = "api:default"
	}

	s.mu.RLock()
	handler := s.chatHandler
	s.mu.RUnlock()

	if handler == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "chat handler not configured"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	response, err := handler(ctx, req.Message, req.SessionKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, ChatResponse{
		Response:   response,
		SessionKey: req.SessionKey,
		Timestamp:  time.Now(),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	status := StatusResponse{
		Status:    "running",
		Timestamp: time.Now(),
	}

	if s.router != nil {
		status.EventSources = s.router.SourceCount()
		status.EventSubscriptions = s.router.SubscriptionCount()
		status.EventRouterRunning = s.router.IsRunning()
	}

	if s.stateMgr != nil {
		status.TrackedUsers = s.stateMgr.UserCount()
	}

	s.mu.RLock()
	status.WebSocketClients = len(s.wsClients)
	s.mu.RUnlock()

	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	if s.stateMgr == nil {
		writeJSON(w, http.StatusOK, UsersResponse{Users: map[string]*state.UserState{}})
		return
	}

	users := s.stateMgr.GetAllUsers()
	if users == nil {
		users = map[string]*state.UserState{}
	}

	writeJSON(w, http.StatusOK, UsersResponse{Users: users})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Kind == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "kind is required"})
		return
	}

	if s.router == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "event router not configured"})
		return
	}

	event := events.NewEvent(req.Kind, "api", req.Priority).
		WithChannel(req.Channel, req.ChatID)
	for k, v := range req.Payload {
		event = event.WithPayload(k, v)
	}

	s.router.Emit(event)

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":   "accepted",
		"event_id": event.ID,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ready := true
	s.mu.RLock()
	if s.chatHandler == nil {
		ready = false
	}
	s.mu.RUnlock()

	if ready {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	} else {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
	}
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
