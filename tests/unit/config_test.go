package unit

import (
	"os"
	"path/filepath"
	"testing"

	"webserver/internal/config"
	"webserver/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager_LoadConfig(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		wantErr    bool
	}{
		{
			name: "valid configuration",
			configData: `{
				"server": {
					"port": 8080,
					"host": "localhost",
					"static_dir": "./static"
				},
				"endpoints": {
					"/api/test": {
						"type": "error",
						"status_code": 500,
						"message": "Test error"
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "invalid port",
			configData: `{
				"server": {
					"port": 70000,
					"host": "localhost",
					"static_dir": "./static"
				},
				"endpoints": {}
			}`,
			wantErr: true,
		},
		{
			name: "invalid JSON",
			configData: `{
				"server": {
					"port": 8080,
					"host": "localhost"
				}
			`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.json")

			err := os.WriteFile(configPath, []byte(tt.configData), 0644)
			require.NoError(t, err)

			// Create config manager
			manager := config.NewManager(configPath)

			// Test loading
			err = manager.LoadConfig()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify config was loaded
				cfg := manager.GetConfig()
				assert.NotNil(t, cfg)
				assert.Equal(t, 8080, cfg.Server.Port)
				assert.Equal(t, "localhost", cfg.Server.Host)
			}
		})
	}
}

func TestConfigManager_UpdateConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	manager := config.NewManager(configPath)

	// Load initial config (should create default)
	err := manager.LoadConfig()
	require.NoError(t, err)

	// Update config
	newConfig := &types.Config{
		Server: types.ServerConfig{
			Port:      9090,
			Host:      "0.0.0.0",
			StaticDir: "./public",
		},
		Endpoints: map[string]types.EndpointConfig{
			"/api/updated": {
				Type:     "delay",
				DelayMs:  1000,
				Response: map[string]interface{}{"status": "updated"},
			},
		},
	}

	err = manager.UpdateConfig(newConfig)
	assert.NoError(t, err)

	// Verify config was updated
	cfg := manager.GetConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "./public", cfg.Server.StaticDir)
	assert.Contains(t, cfg.Endpoints, "/api/updated")

	// Verify file was updated
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}

func TestConfigManager_UpdateEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	manager := config.NewManager(configPath)

	// Load initial config
	err := manager.LoadConfig()
	require.NoError(t, err)

	// Add endpoint
	endpointConfig := types.EndpointConfig{
		Type:            "conditional_error",
		ErrorEveryN:     3,
		StatusCode:      503,
		SuccessResponse: map[string]interface{}{"status": "ok"},
	}

	err = manager.UpdateEndpoint("/api/flaky", endpointConfig)
	assert.NoError(t, err)

	// Verify endpoint was added
	cfg := manager.GetConfig()
	assert.NotNil(t, cfg)
	assert.Contains(t, cfg.Endpoints, "/api/flaky")
	assert.Equal(t, "conditional_error", cfg.Endpoints["/api/flaky"].Type)
	assert.Equal(t, 3, cfg.Endpoints["/api/flaky"].ErrorEveryN)
}

func TestConfigManager_RemoveEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	manager := config.NewManager(configPath)

	// Load initial config
	err := manager.LoadConfig()
	require.NoError(t, err)

	// Get initial endpoints
	cfg := manager.GetConfig()
	initialEndpoints := len(cfg.Endpoints)

	// Find first endpoint to remove
	var pathToRemove string
	for path := range cfg.Endpoints {
		pathToRemove = path
		break
	}

	require.NotEmpty(t, pathToRemove)

	// Remove endpoint
	err = manager.RemoveEndpoint(pathToRemove)
	assert.NoError(t, err)

	// Verify endpoint was removed
	cfg = manager.GetConfig()
	assert.NotNil(t, cfg)
	assert.NotContains(t, cfg.Endpoints, pathToRemove)
	assert.Equal(t, initialEndpoints-1, len(cfg.Endpoints))
}

func TestConfigManager_DefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.json")

	manager := config.NewManager(configPath)

	// Load config (should create default)
	err := manager.LoadConfig()
	assert.NoError(t, err)

	// Verify default config
	cfg := manager.GetConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "./static", cfg.Server.StaticDir)
	assert.Greater(t, len(cfg.Endpoints), 0)

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}
