package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"webserver/pkg/types"
)

// Manager handles configuration loading, validation, and hot reloading
type Manager struct {
	configPath string
	config     *types.Config
	mutex      sync.RWMutex
	watchers   []func(*types.Config)
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
		watchers:   make([]func(*types.Config), 0),
	}
}

// LoadConfig loads the configuration from file
func (m *Manager) LoadConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Create default configuration if file doesn't exist
		defaultConfig := m.createDefaultConfig()
		if err := m.saveConfigToFile(defaultConfig); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		m.config = defaultConfig
		return nil
	}

	// Load existing configuration
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := m.validateConfig(&config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.config = &config
	return nil
}

// GetConfig returns a copy of the current configuration
func (m *Manager) GetConfig() *types.Config {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.config == nil {
		return nil
	}

	// Create a deep copy to avoid race conditions
	configCopy := *m.config
	configCopy.Endpoints = make(map[string]types.EndpointConfig)
	for k, v := range m.config.Endpoints {
		configCopy.Endpoints[k] = v
	}

	return &configCopy
}

// UpdateConfig updates the configuration and saves it to file
func (m *Manager) UpdateConfig(newConfig *types.Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate new configuration
	if err := m.validateConfig(newConfig); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save to file
	if err := m.saveConfigToFile(newConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update in-memory configuration
	m.config = newConfig

	// Notify watchers
	go m.notifyWatchers(newConfig)

	return nil
}

// UpdateEndpoint adds or updates a specific endpoint configuration
func (m *Manager) UpdateEndpoint(path string, endpointConfig types.EndpointConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Validate endpoint configuration
	if err := m.validateEndpointConfig(&endpointConfig); err != nil {
		return fmt.Errorf("invalid endpoint configuration: %w", err)
	}

	// Update endpoint
	if m.config.Endpoints == nil {
		m.config.Endpoints = make(map[string]types.EndpointConfig)
	}
	m.config.Endpoints[path] = endpointConfig

	// Save to file
	if err := m.saveConfigToFile(m.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Notify watchers
	go m.notifyWatchers(m.config)

	return nil
}

// RemoveEndpoint removes an endpoint configuration
func (m *Manager) RemoveEndpoint(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if m.config.Endpoints == nil {
		return fmt.Errorf("endpoint not found")
	}

	delete(m.config.Endpoints, path)

	// Save to file
	if err := m.saveConfigToFile(m.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Notify watchers
	go m.notifyWatchers(m.config)

	return nil
}

// AddWatcher adds a configuration change watcher
func (m *Manager) AddWatcher(watcher func(*types.Config)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.watchers = append(m.watchers, watcher)
}

// createDefaultConfig creates a default configuration
func (m *Manager) createDefaultConfig() *types.Config {
	return &types.Config{
		Server: types.ServerConfig{
			Port:      8080,
			Host:      "0.0.0.0",
			StaticDir: "./static",
		},
		Endpoints: map[string]types.EndpointConfig{
			"/api/error": {
				Type:       "error",
				StatusCode: 500,
				Message:    "Internal Server Error",
			},
			"/api/delay": {
				Type:    "delay",
				DelayMs: 2000,
				Response: map[string]interface{}{
					"message": "Delayed response",
				},
			},
			"/api/flaky": {
				Type:        "conditional_error",
				ErrorEveryN: 3,
				StatusCode:  503,
				SuccessResponse: map[string]interface{}{
					"status": "ok",
				},
			},
		},
	}
}

// validateConfig validates the entire configuration
func (m *Manager) validateConfig(config *types.Config) error {
	// Validate server configuration
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", config.Server.Port)
	}

	if config.Server.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if config.Server.StaticDir == "" {
		return fmt.Errorf("static directory cannot be empty")
	}

	// Validate endpoint configurations
	for path, endpointConfig := range config.Endpoints {
		if path == "" {
			return fmt.Errorf("endpoint path cannot be empty")
		}

		if err := m.validateEndpointConfig(&endpointConfig); err != nil {
			return fmt.Errorf("invalid endpoint '%s': %w", path, err)
		}
	}

	return nil
}

// validateEndpointConfig validates a single endpoint configuration
func (m *Manager) validateEndpointConfig(config *types.EndpointConfig) error {
	switch config.Type {
	case "error":
		if config.StatusCode < 400 || config.StatusCode > 599 {
			return fmt.Errorf("invalid error status code: %d", config.StatusCode)
		}
	case "delay":
		if config.DelayMs < 0 {
			return fmt.Errorf("delay cannot be negative: %d", config.DelayMs)
		}
	case "conditional_error":
		if config.ErrorEveryN < 1 {
			return fmt.Errorf("error_every_n must be at least 1: %d", config.ErrorEveryN)
		}
		if config.StatusCode < 400 || config.StatusCode > 599 {
			return fmt.Errorf("invalid error status code: %d", config.StatusCode)
		}
	case "static":
		// Static endpoints are handled differently
	default:
		return fmt.Errorf("unknown endpoint type: %s", config.Type)
	}

	return nil
}

// saveConfigToFile saves the configuration to file
func (m *Manager) saveConfigToFile(config *types.Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal configuration to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// notifyWatchers notifies all registered watchers of configuration changes
func (m *Manager) notifyWatchers(config *types.Config) {
	for _, watcher := range m.watchers {
		watcher(config)
	}
}

// GetConfigPath returns the path to the configuration file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}
