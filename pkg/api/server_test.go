package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amit-vikramaditya/v1claw/pkg/events"
	"github.com/amit-vikramaditya/v1claw/pkg/state"
)

func newTestServer(t *testing.T) (*Server, *state.Manager) {
	tmpDir := t.TempDir()
	stateMgr := state.NewManager(tmpDir)
	router := events.NewRouter()

	srv := NewServer(Config{Addr: ":0"}, nil, router, stateMgr)
	return srv, stateMgr
}

func TestHealthEndpoint(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "ok", resp["status"])
}

func TestReadyEndpoint_NotReady(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	srv.handleReady(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestReadyEndpoint_Ready(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.SetChatHandler(func(ctx context.Context, msg, sess string) (string, error) {
		return "ok", nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	srv.handleReady(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatusEndpoint(t *testing.T) {
	srv, stateMgr := newTestServer(t)
	stateMgr.SetUserState("u1", "telegram", "1", "alice")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	w := httptest.NewRecorder()
	srv.handleStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp StatusResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "running", resp.Status)
	assert.Equal(t, 1, resp.TrackedUsers)
}

func TestChatEndpoint_Success(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.SetChatHandler(func(ctx context.Context, msg, sess string) (string, error) {
		return "Hello! You said: " + msg, nil
	})

	body, _ := json.Marshal(ChatRequest{Message: "Hi", SessionKey: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.handleChat(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ChatResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "Hello! You said: Hi", resp.Response)
	assert.Equal(t, "test", resp.SessionKey)
}

func TestChatEndpoint_EmptyMessage(t *testing.T) {
	srv, _ := newTestServer(t)

	body, _ := json.Marshal(ChatRequest{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.handleChat(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatEndpoint_NoHandler(t *testing.T) {
	srv, _ := newTestServer(t)

	body, _ := json.Marshal(ChatRequest{Message: "Hi"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.handleChat(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestChatEndpoint_DefaultSessionKey(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.SetChatHandler(func(ctx context.Context, msg, sess string) (string, error) {
		return sess, nil
	})

	body, _ := json.Marshal(ChatRequest{Message: "Hi"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.handleChat(w, req)

	var resp ChatResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "api:default", resp.Response)
}

func TestChatEndpoint_WrongMethod(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat", nil)
	w := httptest.NewRecorder()
	srv.handleChat(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestUsersEndpoint(t *testing.T) {
	srv, stateMgr := newTestServer(t)
	stateMgr.SetUserState("telegram:alice", "telegram", "111", "alice")
	stateMgr.SetUserState("discord:bob", "discord", "222", "bob")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	srv.handleUsers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp UsersResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp.Users, 2)
}

func TestEventsEndpoint(t *testing.T) {
	srv, _ := newTestServer(t)

	ctx := context.Background()
	srv.router.Start(ctx)
	defer srv.router.Stop()

	body, _ := json.Marshal(EventRequest{
		Kind:    "custom",
		Payload: map[string]interface{}{"key": "value"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.handleEvents(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestEventsEndpoint_MissingKind(t *testing.T) {
	srv, _ := newTestServer(t)

	body, _ := json.Marshal(EventRequest{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.handleEvents(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthMiddleware_NoKey(t *testing.T) {
	srv, _ := newTestServer(t)
	// No API key configured — all requests pass
	handler := srv.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_ValidKey(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.apiKey = "secret123"

	handler := srv.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_InvalidKey(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.apiKey = "secret123"

	handler := srv.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_QueryParam(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.apiKey = "secret123"

	handler := srv.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/?api_key=secret123", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewServer_DefaultAddr(t *testing.T) {
	srv := NewServer(Config{}, nil, nil, nil)
	assert.Equal(t, ":18791", srv.addr)
}

func TestNewServer_CustomAddr(t *testing.T) {
	srv := NewServer(Config{Addr: ":9090"}, nil, nil, nil)
	assert.Equal(t, ":9090", srv.addr)
}

func TestStatusEndpoint_WrongMethod(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/status", nil)
	w := httptest.NewRecorder()
	srv.handleStatus(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestServerStartStop(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.addr = ":0" // Random port

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop in time")
	}
}
