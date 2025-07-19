package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"webserver/internal/config"
	"webserver/pkg/types"

	"github.com/gorilla/websocket"
)

// Server represents the configurable web server
type Server struct {
	config          *config.Manager
	configWatcher   *config.Watcher
	httpServer      *http.Server
	stats           *types.ServerStats
	mux             *http.ServeMux
	wsUpgrader      websocket.Upgrader
	wsConnections   map[*websocket.Conn]bool
	wsConnectionsMu sync.RWMutex
	isRunning       bool
	mu              sync.RWMutex

	// Request logging
	requestLog   []types.RequestLogEntry
	requestLogMu sync.RWMutex
	maxLogSize   int
}

// NewServer creates a new configurable web server
func NewServer(configPath string) (*Server, error) {
	configManager := config.NewManager(configPath)
	configWatcher := config.NewWatcher(configManager)

	s := &Server{
		config:        configManager,
		configWatcher: configWatcher,
		stats: &types.ServerStats{
			StartTime: time.Now(),
			Endpoints: make(map[string]*types.EndpointStats),
		},
		mux:           http.NewServeMux(),
		wsUpgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		wsConnections: make(map[*websocket.Conn]bool),
		requestLog:    make([]types.RequestLogEntry, 0),
		maxLogSize:    1000, // Keep last 1000 requests
	}

	// Load initial configuration
	if err := s.config.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set up configuration change watcher
	s.config.AddWatcher(s.onConfigChange)

	// Set up routes
	s.setupRoutes()

	return s, nil
}

// Start starts the web server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	currentConfig := s.config.GetConfig()
	if currentConfig == nil {
		return fmt.Errorf("no configuration loaded")
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", currentConfig.Server.Host, currentConfig.Server.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.logRequestMiddleware(s.mux), // Wrap with logging middleware
	}

	// Start configuration file watcher
	if err := s.configWatcher.Start(); err != nil {
		return fmt.Errorf("failed to start config watcher: %w", err)
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	s.isRunning = true
	log.Printf("Server started successfully on %s", addr)
	return nil
}

// Stop stops the web server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	// Stop configuration watcher
	s.configWatcher.Stop()

	// Close all WebSocket connections
	s.wsConnectionsMu.Lock()
	for conn := range s.wsConnections {
		conn.Close()
	}
	s.wsConnections = make(map[*websocket.Conn]bool)
	s.wsConnectionsMu.Unlock()

	// Shutdown HTTP server
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	s.isRunning = false
	log.Println("Server stopped successfully")
	return nil
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetStats returns the current server statistics
func (s *Server) GetStats() types.ServerStats {
	return s.stats.GetAllStats()
}

// setupRoutes sets up the HTTP routes
func (s *Server) setupRoutes() {
	// Configuration management endpoint
	s.mux.HandleFunc("/config", s.handleConfig)

	// WebSocket endpoint for TUI
	s.mux.HandleFunc("/ws", s.handleWebSocket)

	// Statistics endpoint
	s.mux.HandleFunc("/stats", s.handleStats)

	// Request log endpoint
	s.mux.HandleFunc("/requestlog", s.handleRequestLog)

	// Catch-all handler for dynamic endpoints and static files
	s.mux.HandleFunc("/", s.handleRequest)
}

// onConfigChange handles configuration changes
func (s *Server) onConfigChange(newConfig *types.Config) {
	log.Println("Configuration changed, updating server...")

	// Check if server address changed
	currentConfig := s.config.GetConfig()
	if currentConfig.Server.Host != newConfig.Server.Host ||
		currentConfig.Server.Port != newConfig.Server.Port {
		log.Println("Server address changed, restart required")
		// In a production system, you might want to handle this more gracefully
	}

	// Broadcast configuration change to WebSocket clients
	s.broadcastToWebSockets(types.TUIMessage{
		Type:      "config_updated",
		Timestamp: time.Now(),
		Data:      newConfig,
	})

	log.Println("Configuration updated successfully")
}

// addWebSocketConnection adds a new WebSocket connection
func (s *Server) addWebSocketConnection(conn *websocket.Conn) {
	s.wsConnectionsMu.Lock()
	defer s.wsConnectionsMu.Unlock()
	s.wsConnections[conn] = true
}

// removeWebSocketConnection removes a WebSocket connection
func (s *Server) removeWebSocketConnection(conn *websocket.Conn) {
	s.wsConnectionsMu.Lock()
	defer s.wsConnectionsMu.Unlock()
	delete(s.wsConnections, conn)
}

// broadcastToWebSockets broadcasts a message to all connected WebSocket clients
func (s *Server) broadcastToWebSockets(message types.TUIMessage) {
	s.wsConnectionsMu.RLock()
	defer s.wsConnectionsMu.RUnlock()

	for conn := range s.wsConnections {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Failed to send WebSocket message: %v", err)
			// Remove bad connection
			delete(s.wsConnections, conn)
			conn.Close()
		}
	}
}

// ensureStaticDir ensures the static directory exists
func (s *Server) ensureStaticDir(staticDir string) error {
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.MkdirAll(staticDir, 0755); err != nil {
			return fmt.Errorf("failed to create static directory: %w", err)
		}

		// Create a default index.html
		indexContent := `<!DOCTYPE html>
<html>
<head>
    		<title>WebServer</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        h1 { color: #333; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .code { background: #eeeeee; padding: 5px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="container">
        		<h1>WebServer Configurable Web Server</h1>
		<p>Welcome to the WebServer! This server is running and ready to handle requests.</p>
        
        <h2>Available Endpoints</h2>
        <div class="endpoint">
            <strong>GET /config</strong> - Get current configuration
        </div>
        <div class="endpoint">
            <strong>PUT /config</strong> - Update configuration
        </div>
        <div class="endpoint">
            <strong>GET /stats</strong> - Get server statistics
        </div>
        <div class="endpoint">
            <strong>GET /ws</strong> - WebSocket endpoint for TUI
        </div>
        
        <h2>Testing Dynamic Endpoints</h2>
        <p>Try these default endpoints to test the dynamic behavior:</p>
        <div class="endpoint">
            <strong>GET <a href="/api/error">/api/error</a></strong> - Returns a 500 error
        </div>
        <div class="endpoint">
            <strong>GET <a href="/api/delay">/api/delay</a></strong> - Returns a delayed response (2 seconds)
        </div>
        <div class="endpoint">
            <strong>GET <a href="/api/flaky">/api/flaky</a></strong> - Returns an error every 3rd request
        </div>
        
        <h2>Configuration</h2>
        <p>The server configuration is hot-reloadable. Modify the configuration file or use the <span class="code">/config</span> endpoint to update settings.</p>
    </div>
</body>
</html>`

		indexPath := fmt.Sprintf("%s/index.html", staticDir)
		if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			return fmt.Errorf("failed to create index.html: %w", err)
		}

		log.Printf("Created static directory and default index.html at %s", staticDir)
	}
	return nil
}

// GetRequestLog returns a copy of the current request log
func (s *Server) GetRequestLog() []types.RequestLogEntry {
	s.requestLogMu.RLock()
	defer s.requestLogMu.RUnlock()

	// Return a copy to avoid race conditions
	logCopy := make([]types.RequestLogEntry, len(s.requestLog))
	copy(logCopy, s.requestLog)
	return logCopy
}

// addToRequestLog adds a request entry to the stored request log
func (s *Server) addToRequestLog(entry types.RequestLogEntry) {
	s.requestLogMu.Lock()
	defer s.requestLogMu.Unlock()

	// Add to beginning of slice (newest first)
	s.requestLog = append([]types.RequestLogEntry{entry}, s.requestLog...)

	// Trim to max size
	if len(s.requestLog) > s.maxLogSize {
		s.requestLog = s.requestLog[:s.maxLogSize]
	}
}

// logRequestMiddleware wraps handlers to log all requests
func (s *Server) logRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Create a response writer that captures the status code
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log the request (this calls the existing logRequest method)
		s.logRequest(r)

		// Add to stored request log and broadcast to WebSocket clients
		duration := time.Since(startTime)
		entry := types.RequestLogEntry{
			Timestamp:  startTime,
			Method:     r.Method,
			Path:       r.URL.RequestURI(), // Use full request URI including query parameters
			StatusCode: rw.statusCode,
			Duration:   duration.Milliseconds(),
			RemoteAddr: r.RemoteAddr,
		}

		s.addToRequestLog(entry)
		s.broadcastToWebSockets(types.TUIMessage{
			Type: "request_log",
			Data: entry,
		})
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
