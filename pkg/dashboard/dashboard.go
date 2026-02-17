package dashboard

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
)

//go:embed templates/*.html
var templateFS embed.FS

// Config holds dashboard configuration.
type Config struct {
	Enabled bool   `json:"enabled"`
	Addr    string `json:"addr"` // Default ":18792"
	Title   string `json:"title"`
}

// StatusData provides real-time status for the dashboard.
type StatusData struct {
	Version            string    `json:"version"`
	Uptime             string    `json:"uptime"`
	Status             string    `json:"status"`
	ActiveChannels     []string  `json:"active_channels"`
	TrackedUsers       int       `json:"tracked_users"`
	EventSources       int       `json:"event_sources"`
	EventSubscriptions int       `json:"event_subscriptions"`
	PendingJobs        int       `json:"pending_jobs"`
	KnowledgeDocs      int       `json:"knowledge_docs"`
	ConnectedDevices   int       `json:"connected_devices"`
	WebSocketClients   int       `json:"websocket_clients"`
	Timestamp          time.Time `json:"timestamp"`
}

// StatusProvider is called to get current system status.
type StatusProvider func() StatusData

// Server serves the V1 web dashboard.
type Server struct {
	mu             sync.RWMutex
	config         Config
	httpServer     *http.Server
	templates      *template.Template
	statusProvider StatusProvider
	startTime      time.Time
}

// NewServer creates a new dashboard server.
func NewServer(cfg Config) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":18792"
	}
	if cfg.Title == "" {
		cfg.Title = "V1 Dashboard"
	}

	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		logger.ErrorC("dashboard", fmt.Sprintf("Failed to parse templates: %v", err))
		// Create a minimal template as fallback.
		tmpl = template.Must(template.New("index.html").Parse(fallbackTemplate))
	}

	return &Server{
		config:    cfg,
		templates: tmpl,
		startTime: time.Now(),
	}
}

// SetStatusProvider sets the function that provides status data.
func (s *Server) SetStatusProvider(provider StatusProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statusProvider = provider
}

// Start begins serving the dashboard.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/status", s.handleStatus)

	s.httpServer = &http.Server{
		Addr:         s.config.Addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.InfoC("dashboard", fmt.Sprintf("Dashboard starting on %s", s.config.Addr))

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
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := s.getStatus()
	s.templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Title":  s.config.Title,
		"Status": data,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	data := s.getStatus()

	// Return as HTML fragment for htmx polling.
	s.templates.ExecuteTemplate(w, "status_fragment.html", data)
}

func (s *Server) getStatus() StatusData {
	s.mu.RLock()
	provider := s.statusProvider
	s.mu.RUnlock()

	if provider != nil {
		data := provider()
		data.Uptime = time.Since(s.startTime).Truncate(time.Second).String()
		data.Timestamp = time.Now()
		return data
	}

	return StatusData{
		Status:    "running",
		Uptime:    time.Since(s.startTime).Truncate(time.Second).String(),
		Timestamp: time.Now(),
	}
}

const fallbackTemplate = `<!DOCTYPE html>
<html><head><title>V1 Dashboard</title></head>
<body><h1>V1 Dashboard</h1><p>Status: {{.Status.Status}}</p><p>Uptime: {{.Status.Uptime}}</p></body>
</html>`
