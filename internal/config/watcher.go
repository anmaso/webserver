package config

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher handles file system watching for configuration hot reloading
type Watcher struct {
	manager   *Manager
	watcher   *fsnotify.Watcher
	stopChan  chan struct{}
	isRunning bool
	mutex     sync.Mutex
}

// NewWatcher creates a new configuration file watcher
func NewWatcher(manager *Manager) *Watcher {
	return &Watcher{
		manager:  manager,
		stopChan: make(chan struct{}),
	}
}

// Start starts watching the configuration file for changes
func (w *Watcher) Start() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isRunning {
		return nil
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	w.watcher = watcher
	w.isRunning = true

	// Watch the configuration file and its directory
	configPath := w.manager.GetConfigPath()
	configDir := filepath.Dir(configPath)

	// Add directory to watcher (needed for file creation/deletion)
	if err := w.watcher.Add(configDir); err != nil {
		w.watcher.Close()
		w.isRunning = false
		return err
	}

	// Start the watching goroutine
	go w.watch()

	log.Printf("Started configuration file watcher for: %s", configPath)
	return nil
}

// Stop stops the file watcher
func (w *Watcher) Stop() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isRunning {
		return
	}

	close(w.stopChan)
	w.watcher.Close()
	w.isRunning = false
	log.Println("Stopped configuration file watcher")
}

// IsRunning returns whether the watcher is currently running
func (w *Watcher) IsRunning() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.isRunning
}

// watch is the main watching loop
func (w *Watcher) watch() {
	configPath := w.manager.GetConfigPath()
	configFileName := filepath.Base(configPath)

	// Debounce file changes to avoid multiple reloads
	var lastReload time.Time
	debounceInterval := 500 * time.Millisecond

	for {
		select {
		case <-w.stopChan:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Check if the event is for our configuration file
			if filepath.Base(event.Name) != configFileName {
				continue
			}

			// Debounce rapid file changes
			if time.Since(lastReload) < debounceInterval {
				continue
			}

			// Handle different event types
			switch {
			case event.Op&fsnotify.Write == fsnotify.Write:
				log.Printf("Configuration file modified: %s", event.Name)
				w.reloadConfig()
				lastReload = time.Now()
			case event.Op&fsnotify.Create == fsnotify.Create:
				log.Printf("Configuration file created: %s", event.Name)
				w.reloadConfig()
				lastReload = time.Now()
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				log.Printf("Configuration file removed: %s", event.Name)
				// Could handle this by creating a default config
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				log.Printf("Configuration file renamed: %s", event.Name)
				// Could handle this by re-adding the watcher
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

// reloadConfig reloads the configuration from file
func (w *Watcher) reloadConfig() {
	// Add a small delay to ensure file write is complete
	time.Sleep(100 * time.Millisecond)

	if err := w.manager.LoadConfig(); err != nil {
		log.Printf("Failed to reload configuration: %v", err)
	} else {
		log.Println("Configuration reloaded successfully")
	}
}
