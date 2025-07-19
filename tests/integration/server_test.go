package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"webserver/internal/server"
	"webserver/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create and start server
	srv, err := server.NewServer(configPath)
	require.NoError(t, err)

	err = srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	baseURL := "http://localhost:8080"

	// Test configuration endpoint
	t.Run("GET /config", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/config")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var config types.Config
		err = json.NewDecoder(resp.Body).Decode(&config)
		require.NoError(t, err)

		assert.Equal(t, 8080, config.Server.Port)
		assert.Equal(t, "0.0.0.0", config.Server.Host)
		assert.Greater(t, len(config.Endpoints), 0)
	})

	// Test statistics endpoint
	t.Run("GET /stats", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/stats")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var stats types.ServerStats
		err = json.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)

		assert.False(t, stats.StartTime.IsZero())
	})

	// Test dynamic endpoints
	t.Run("Dynamic endpoints", func(t *testing.T) {
		// Test error endpoint
		t.Run("Error endpoint", func(t *testing.T) {
			resp, err := http.Get(baseURL + "/api/error")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "error")
		})

		// Test delay endpoint
		t.Run("Delay endpoint", func(t *testing.T) {
			start := time.Now()
			resp, err := http.Get(baseURL + "/api/delay")
			duration := time.Since(start)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Greater(t, duration, time.Second) // Should be delayed

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "message")
		})

		// Test conditional error endpoint
		t.Run("Conditional error endpoint", func(t *testing.T) {
			successCount := 0
			errorCount := 0

			// Make several requests to test the pattern
			for i := 0; i < 6; i++ {
				resp, err := http.Get(baseURL + "/api/flaky")
				require.NoError(t, err)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					successCount++
				} else {
					errorCount++
				}
			}

			// Should have both successes and errors
			assert.Greater(t, successCount, 0)
			assert.Greater(t, errorCount, 0)
		})
	})

	// Test static file serving
	t.Run("Static file serving", func(t *testing.T) {
		// Test root path (should serve index.html)
		resp, err := http.Get(baseURL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Contains(t, string(body), "WebServer")
		assert.Contains(t, string(body), "html")
	})

	// Test configuration update
	t.Run("Configuration update", func(t *testing.T) {
		// Add new endpoint
		newEndpoint := map[string]interface{}{
			"path": "/api/test",
			"config": map[string]interface{}{
				"type":        "error",
				"status_code": 404,
				"message":     "Test endpoint",
			},
		}

		body, err := json.Marshal(newEndpoint)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/config", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Wait for configuration to be applied
		time.Sleep(100 * time.Millisecond)

		// Test new endpoint
		resp, err = http.Get(baseURL + "/api/test")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestServerConfigurationPersistence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create initial configuration
	initialConfig := types.Config{
		Server: types.ServerConfig{
			Port:      8081,
			Host:      "127.0.0.1",
			StaticDir: "./static",
		},
		Endpoints: map[string]types.EndpointConfig{
			"/api/persist": {
				Type:       "error",
				StatusCode: 503,
				Message:    "Persistence test",
			},
		},
	}

	configData, err := json.MarshalIndent(initialConfig, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(t, err)

	// Create and start server
	srv, err := server.NewServer(configPath)
	require.NoError(t, err)

	err = srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	baseURL := "http://127.0.0.1:8081"

	// Test that persisted configuration is loaded
	t.Run("Persisted config loaded", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/persist")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Persistence test", response["error"])
	})

	// Test configuration persistence after update
	t.Run("Configuration persists after update", func(t *testing.T) {
		// Update configuration
		newEndpoint := map[string]interface{}{
			"path": "/api/new",
			"config": map[string]interface{}{
				"type":     "delay",
				"delay_ms": 500,
				"response": map[string]interface{}{
					"status": "delayed",
				},
			},
		}

		body, err := json.Marshal(newEndpoint)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/config", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Wait for configuration to be applied and persisted
		time.Sleep(200 * time.Millisecond)

		// Verify configuration was persisted to file
		configData, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var persistedConfig types.Config
		err = json.Unmarshal(configData, &persistedConfig)
		require.NoError(t, err)

		assert.Contains(t, persistedConfig.Endpoints, "/api/new")
		assert.Equal(t, "delay", persistedConfig.Endpoints["/api/new"].Type)
		assert.Equal(t, 500, persistedConfig.Endpoints["/api/new"].DelayMs)
	})
}

func TestServerStatisticsTracking(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	srv, err := server.NewServer(configPath)
	require.NoError(t, err)

	err = srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	time.Sleep(100 * time.Millisecond)

	baseURL := "http://localhost:8080"

	t.Run("Statistics tracking", func(t *testing.T) {
		// Make several requests
		for i := 0; i < 5; i++ {
			resp, err := http.Get(baseURL + "/api/error")
			require.NoError(t, err)
			resp.Body.Close()
		}

		// Wait for statistics to be processed
		time.Sleep(100 * time.Millisecond)

		// Check statistics
		resp, err := http.Get(baseURL + "/stats")
		require.NoError(t, err)
		defer resp.Body.Close()

		var stats types.ServerStats
		err = json.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)

		// Verify statistics were recorded
		assert.Greater(t, stats.RequestCount, int64(0))
		assert.Greater(t, stats.ErrorCount, int64(0))
		assert.Contains(t, stats.Endpoints, "/api/error")

		errorStats := stats.Endpoints["/api/error"]
		assert.Greater(t, errorStats.RequestCount, int64(0))
		assert.Greater(t, errorStats.ErrorCount, int64(0))
		assert.Contains(t, errorStats.StatusCodes, 500)
	})
}
