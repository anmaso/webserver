package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"webserver/pkg/types"

	"github.com/gorilla/websocket"
)

// handleConfig handles configuration management endpoints
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		s.stats.RecordRequest("/config", time.Since(start), http.StatusOK)
	}()

	switch r.Method {
	case http.MethodGet:
		s.handleGetConfig(w, r)
	case http.MethodPut:
		s.handleUpdateConfig(w, r)
	case http.MethodPost:
		s.handleAddEndpoint(w, r)
	case http.MethodDelete:
		s.handleRemoveEndpoint(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetConfig returns the current configuration
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	config := s.config.GetConfig()
	if config == nil {
		http.Error(w, "Configuration not loaded", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// handleUpdateConfig updates the entire configuration
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig types.Config
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.config.UpdateConfig(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update configuration: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Configuration updated"})
}

// handleAddEndpoint adds or updates a specific endpoint
func (s *Server) handleAddEndpoint(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Path   string               `json:"path"`
		Config types.EndpointConfig `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	if err := s.config.UpdateEndpoint(request.Path, request.Config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add endpoint: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Endpoint added"})
}

// handleRemoveEndpoint removes an endpoint
func (s *Server) handleRemoveEndpoint(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	if err := s.config.RemoveEndpoint(path); err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove endpoint: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Endpoint removed"})
}

// handleStats returns server statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		s.stats.RecordRequest("/stats", time.Since(start), http.StatusOK)
	}()

	stats := s.stats.GetAllStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleWebSocket handles WebSocket connections for TUI
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Add connection to active connections
	s.addWebSocketConnection(conn)
	defer s.removeWebSocketConnection(conn)

	log.Printf("New WebSocket connection from %s", r.RemoteAddr)

	// Send initial data
	s.sendInitialData(conn)

	// Handle incoming messages
	for {
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		s.handleWebSocketMessage(conn, message)
	}
}

// sendInitialData sends initial configuration and statistics to new WebSocket client
func (s *Server) sendInitialData(conn *websocket.Conn) {
	// Send current configuration
	config := s.config.GetConfig()
	if config != nil {
		conn.WriteJSON(types.TUIMessage{
			Type:      "config",
			Timestamp: time.Now(),
			Data:      config,
		})
	}

	// Send current statistics
	stats := s.stats.GetAllStats()
	conn.WriteJSON(types.TUIMessage{
		Type:      "stats",
		Timestamp: time.Now(),
		Data:      stats,
	})
}

// handleWebSocketMessage handles incoming WebSocket messages
func (s *Server) handleWebSocketMessage(conn *websocket.Conn, message map[string]interface{}) {
	msgType, ok := message["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "get_config":
		config := s.config.GetConfig()
		conn.WriteJSON(types.TUIMessage{
			Type:      "config",
			Timestamp: time.Now(),
			Data:      config,
		})
	case "get_stats":
		stats := s.stats.GetAllStats()
		conn.WriteJSON(types.TUIMessage{
			Type:      "stats",
			Timestamp: time.Now(),
			Data:      stats,
		})
	}
}

// handleRequest handles all other requests (dynamic endpoints and static files)
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	config := s.config.GetConfig()
	if config == nil {
		http.Error(w, "Server configuration not loaded", http.StatusInternalServerError)
		s.stats.RecordRequest(r.URL.Path, time.Since(start), http.StatusInternalServerError)
		return
	}

	// Note: Request logging is now handled by middleware to avoid duplication

	// Check if this is a configured dynamic endpoint
	if endpointConfig, exists := config.Endpoints[r.URL.Path]; exists {
		s.handleDynamicEndpoint(w, r, endpointConfig)
		return
	}

	// Handle static file serving
	s.handleStaticFile(w, r, config.Server.StaticDir)
}

// handleDynamicEndpoint handles configured dynamic endpoints
func (s *Server) handleDynamicEndpoint(w http.ResponseWriter, r *http.Request, config types.EndpointConfig) {
	start := time.Now()
	endpointStats := s.stats.GetEndpointStats(r.URL.Path)

	var statusCode int
	var responseData interface{}

	switch config.Type {
	case "error":
		statusCode = config.StatusCode
		responseData = map[string]string{"error": config.Message}

	case "delay":
		if config.DelayMs > 0 {
			time.Sleep(time.Duration(config.DelayMs) * time.Millisecond)
		}
		statusCode = http.StatusOK
		responseData = config.Response

	case "conditional_error":
		endpointStats.IncrementConditionalCount()
		count := endpointStats.GetConditionalCount()

		if count%int64(config.ErrorEveryN) == 0 {
			statusCode = config.StatusCode
			responseData = map[string]string{"error": "Conditional error triggered"}
		} else {
			statusCode = http.StatusOK
			responseData = config.SuccessResponse
		}

	default:
		statusCode = http.StatusInternalServerError
		responseData = map[string]string{"error": "Unknown endpoint type"}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(responseData)

	// Record statistics
	s.stats.RecordRequest(r.URL.Path, time.Since(start), statusCode)

	// Note: Request logging is now handled by middleware to avoid duplication
}

// handleStaticFile serves static files
func (s *Server) handleStaticFile(w http.ResponseWriter, r *http.Request, staticDir string) {
	start := time.Now()

	// Ensure static directory exists
	if err := s.ensureStaticDir(staticDir); err != nil {
		log.Printf("Failed to ensure static directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		s.stats.RecordRequest(r.URL.Path, time.Since(start), http.StatusInternalServerError)
		return
	}

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(r.URL.Path)
	if cleanPath == "/" {
		cleanPath = "/index.html"
	}

	// Build full file path
	filePath := filepath.Join(staticDir, cleanPath)

	// Check if file exists and is within static directory
	absStaticDir, err := filepath.Abs(staticDir)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		s.stats.RecordRequest(r.URL.Path, time.Since(start), http.StatusInternalServerError)
		return
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		s.stats.RecordRequest(r.URL.Path, time.Since(start), http.StatusInternalServerError)
		return
	}

	if !strings.HasPrefix(absFilePath, absStaticDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		s.stats.RecordRequest(r.URL.Path, time.Since(start), http.StatusForbidden)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)

	// Record statistics (assume success, ServeFile handles errors)
	s.stats.RecordRequest(r.URL.Path, time.Since(start), http.StatusOK)
}

// logRequest logs the incoming request
func (s *Server) logRequest(r *http.Request) {
	log.Printf("%s %s %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
}

// broadcastRequestLog broadcasts request information to WebSocket clients
func (s *Server) broadcastRequestLog(r *http.Request, statusCode int, duration time.Duration) {
	logEntry := types.RequestLogEntry{
		Timestamp:  time.Now(),
		Method:     r.Method,
		Path:       r.URL.RequestURI(), // Use full request URI including query parameters
		StatusCode: statusCode,
		Duration:   duration.Milliseconds(),
		RemoteAddr: r.RemoteAddr,
	}

	s.broadcastToWebSockets(types.TUIMessage{
		Type:      "request_log",
		Timestamp: time.Now(),
		Data:      logEntry,
	})
}

// handleRequestLog serves the current request log
func (s *Server) handleRequestLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	requestLog := s.GetRequestLog()
	if err := json.NewEncoder(w).Encode(requestLog); err != nil {
		log.Printf("Failed to encode request log: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
