package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"webserver/internal/server"
	"webserver/internal/tui"
)

func main() {
	var (
		configPath = flag.String("config", "configs/default.json", "Path to configuration file")
		client     = flag.Bool("client", false, "Run in client mode (TUI)")
		serverURL  = flag.String("server", "ws://localhost:8080/ws", "WebSocket server URL (client mode only)")
		help       = flag.Bool("help", false, "Show help message")
		version    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	if *client {
		runClient(*serverURL)
	} else {
		runServer(*configPath)
	}
}

func runServer(configPath string) {
	log.Println("Starting webserver...")

	// Create and start server
	srv, err := server.NewServer(configPath)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Server is running. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("Shutting down server...")
	if err := srv.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Server stopped.")
}

func runClient(serverURL string) {
	log.Printf("Starting webserver client, connecting to: %s", serverURL)

	if err := tui.RunTUI(serverURL); err != nil {
		log.Fatalf("Failed to start TUI: %v", err)
	}
}

func showHelp() {
	fmt.Println("WebServer - Configurable Web Server")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  webserver [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -config string")
	fmt.Println("        Path to configuration file (default: configs/default.json)")
	fmt.Println("  -client")
	fmt.Println("        Run in client mode (TUI)")
	fmt.Println("  -server string")
	fmt.Println("        WebSocket server URL for client mode (default: ws://localhost:8080/ws)")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("  -version")
	fmt.Println("        Show version information")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Start server with default configuration")
	fmt.Println("  webserver")
	fmt.Println()
	fmt.Println("  # Start server with custom configuration")
	fmt.Println("  webserver -config /path/to/config.json")
	fmt.Println()
	fmt.Println("  # Run client (TUI) to connect to local server")
	fmt.Println("  webserver --client")
	fmt.Println()
	fmt.Println("  # Run client (TUI) to connect to remote server")
	fmt.Println("  webserver --client -server ws://example.com:8080/ws")
	fmt.Println()
	fmt.Println("SERVER FEATURES:")
	fmt.Println("  - Configurable static file serving")
	fmt.Println("  - Dynamic endpoint responses (errors, delays, conditional errors)")
	fmt.Println("  - Hot configuration reloading")
	fmt.Println("  - Real-time statistics tracking")
	fmt.Println("  - WebSocket API for TUI client")
	fmt.Println("  - RESTful configuration management")
	fmt.Println()
	fmt.Println("CLIENT FEATURES:")
	fmt.Println("  - Real-time server monitoring")
	fmt.Println("  - Configuration viewing and management")
	fmt.Println("  - Statistics dashboard")
	fmt.Println("  - Request log streaming")
	fmt.Println("  - Attractive terminal interface")
	fmt.Println()
	fmt.Println("CONFIGURATION:")
	fmt.Println("  The server uses JSON configuration files with the following structure:")
	fmt.Println("  {")
	fmt.Println("    \"server\": {")
	fmt.Println("      \"port\": 8080,")
	fmt.Println("      \"host\": \"0.0.0.0\",")
	fmt.Println("      \"static_dir\": \"./static\"")
	fmt.Println("    },")
	fmt.Println("    \"endpoints\": {")
	fmt.Println("      \"/api/error\": {")
	fmt.Println("        \"type\": \"error\",")
	fmt.Println("        \"status_code\": 500,")
	fmt.Println("        \"message\": \"Internal Server Error\"")
	fmt.Println("      }")
	fmt.Println("    }")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("API ENDPOINTS:")
	fmt.Println("  GET    /config      - Get current configuration")
	fmt.Println("  PUT    /config      - Update entire configuration")
	fmt.Println("  POST   /config      - Add/update endpoint")
	fmt.Println("  DELETE /config      - Remove endpoint")
	fmt.Println("  GET    /stats       - Get server statistics")
	fmt.Println("  GET    /ws          - WebSocket connection for TUI")
	fmt.Println()
	fmt.Println("CLIENT KEYBOARD SHORTCUTS:")
	fmt.Println("  Tab/Shift+Tab    - Switch between tabs")
	fmt.Println("  R                - Refresh data")
	fmt.Println("  Q/Ctrl+C         - Quit")
	fmt.Println()
}

func showVersion() {
	fmt.Println("WebServer Configurable Web Server")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Built with Go")
	fmt.Println()
	fmt.Println("Server Features:")
	fmt.Println("  ✓ Static file serving")
	fmt.Println("  ✓ Dynamic endpoint configuration")
	fmt.Println("  ✓ Hot configuration reloading")
	fmt.Println("  ✓ Real-time statistics")
	fmt.Println("  ✓ WebSocket TUI support")
	fmt.Println("  ✓ RESTful configuration API")
	fmt.Println()
	fmt.Println("Client Features:")
	fmt.Println("  ✓ Real-time server monitoring")
	fmt.Println("  ✓ Configuration viewing")
	fmt.Println("  ✓ Statistics dashboard")
	fmt.Println("  ✓ Request log streaming")
	fmt.Println("  ✓ Attractive terminal interface")
	fmt.Println("  ✓ WebSocket connectivity")
}
