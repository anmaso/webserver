package types

import (
	"sync"
	"time"
)

// ServerConfig represents the main server configuration
type ServerConfig struct {
	Port      int    `json:"port"`
	Host      string `json:"host"`
	StaticDir string `json:"static_dir"`
}

// EndpointConfig represents configuration for a single endpoint
type EndpointConfig struct {
	Type           string                 `json:"type"`
	StatusCode     int                    `json:"status_code,omitempty"`
	Message        string                 `json:"message,omitempty"`
	DelayMs        int                    `json:"delay_ms,omitempty"`
	Response       map[string]interface{} `json:"response,omitempty"`
	ErrorEveryN    int                    `json:"error_every_n,omitempty"`
	SuccessResponse map[string]interface{} `json:"success_response,omitempty"`
}

// Config represents the complete server configuration
type Config struct {
	Server    ServerConfig              `json:"server"`
	Endpoints map[string]EndpointConfig `json:"endpoints"`
}

// EndpointStats represents statistics for a single endpoint
type EndpointStats struct {
	Path            string             `json:"path"`
	RequestCount    int64              `json:"request_count"`
	ErrorCount      int64              `json:"error_count"`
	TotalTimeMs     int64              `json:"total_time_ms"`
	MinTimeMs       int64              `json:"min_time_ms"`
	MaxTimeMs       int64              `json:"max_time_ms"`
	StatusCodes     map[int]int64      `json:"status_codes"`
	FirstRequest    time.Time          `json:"first_request"`
	LastRequest     time.Time          `json:"last_request"`
	ConditionalCount int64             `json:"conditional_count"` // For N-request pattern tracking
	mutex           sync.RWMutex       `json:"-"`
}

// ServerStats represents overall server statistics
type ServerStats struct {
	StartTime     time.Time                `json:"start_time"`
	RequestCount  int64                    `json:"total_requests"`
	ErrorCount    int64                    `json:"total_errors"`
	Endpoints     map[string]*EndpointStats `json:"endpoints"`
	mutex         sync.RWMutex             `json:"-"`
}

// TUIMessage represents messages sent to the TUI client
type TUIMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// RequestLogEntry represents a single request log entry
type RequestLogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	Duration   int64     `json:"duration_ms"`
	RemoteAddr string    `json:"remote_addr"`
}

// ConfigUpdateRequest represents a request to update configuration
type ConfigUpdateRequest struct {
	Operation string      `json:"operation"` // "set", "add", "remove"
	Path      string      `json:"path"`      // endpoint path for endpoint operations
	Config    interface{} `json:"config"`    // new configuration data
}

// Methods for EndpointStats
func (es *EndpointStats) RecordRequest(duration time.Duration, statusCode int) {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	
	now := time.Now()
	durationMs := duration.Milliseconds()
	
	es.RequestCount++
	es.TotalTimeMs += durationMs
	
	if statusCode >= 400 {
		es.ErrorCount++
	}
	
	if es.MinTimeMs == 0 || durationMs < es.MinTimeMs {
		es.MinTimeMs = durationMs
	}
	
	if durationMs > es.MaxTimeMs {
		es.MaxTimeMs = durationMs
	}
	
	if es.StatusCodes == nil {
		es.StatusCodes = make(map[int]int64)
	}
	es.StatusCodes[statusCode]++
	
	if es.FirstRequest.IsZero() {
		es.FirstRequest = now
	}
	es.LastRequest = now
}

func (es *EndpointStats) IncrementConditionalCount() {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	es.ConditionalCount++
}

func (es *EndpointStats) GetConditionalCount() int64 {
	es.mutex.RLock()
	defer es.mutex.RUnlock()
	return es.ConditionalCount
}

func (es *EndpointStats) GetStats() EndpointStats {
	es.mutex.RLock()
	defer es.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	stats := EndpointStats{
		Path:             es.Path,
		RequestCount:     es.RequestCount,
		ErrorCount:       es.ErrorCount,
		TotalTimeMs:      es.TotalTimeMs,
		MinTimeMs:        es.MinTimeMs,
		MaxTimeMs:        es.MaxTimeMs,
		StatusCodes:      make(map[int]int64),
		FirstRequest:     es.FirstRequest,
		LastRequest:      es.LastRequest,
		ConditionalCount: es.ConditionalCount,
	}
	
	for code, count := range es.StatusCodes {
		stats.StatusCodes[code] = count
	}
	
	return stats
}

// Methods for ServerStats
func (ss *ServerStats) GetEndpointStats(path string) *EndpointStats {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	
	if ss.Endpoints == nil {
		ss.Endpoints = make(map[string]*EndpointStats)
	}
	
	if _, exists := ss.Endpoints[path]; !exists {
		ss.Endpoints[path] = &EndpointStats{
			Path:        path,
			StatusCodes: make(map[int]int64),
		}
	}
	
	return ss.Endpoints[path]
}

func (ss *ServerStats) RecordRequest(path string, duration time.Duration, statusCode int) {
	ss.mutex.Lock()
	ss.RequestCount++
	if statusCode >= 400 {
		ss.ErrorCount++
	}
	ss.mutex.Unlock()
	
	endpointStats := ss.GetEndpointStats(path)
	endpointStats.RecordRequest(duration, statusCode)
}

func (ss *ServerStats) GetAllStats() ServerStats {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	
	stats := ServerStats{
		StartTime:    ss.StartTime,
		RequestCount: ss.RequestCount,
		ErrorCount:   ss.ErrorCount,
		Endpoints:    make(map[string]*EndpointStats),
	}
	
	for path, endpointStats := range ss.Endpoints {
		endpointStatsCopy := endpointStats.GetStats()
		stats.Endpoints[path] = &endpointStatsCopy
	}
	
	return stats
} 