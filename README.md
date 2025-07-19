# WebServer - Configurable Web Server

A highly configurable web server written in Go that serves static files, provides dynamic response generation, and includes real-time monitoring through a beautiful terminal user interface.

## Features

- ✨ **Configurable Static File Serving** - Serve static files from any directory
- 🔄 **Dynamic Response Generation** - Generate responses based on configuration
  - Error responses with custom status codes
  - Delayed responses with configurable delays
  - Conditional error responses (error every N requests)
- 🔥 **Hot Configuration Reloading** - Changes take effect immediately
- 🌐 **RESTful Configuration API** - Manage configuration via HTTP endpoints
- 📊 **Real-time Statistics** - Track requests, errors, and performance metrics
- 🖥️ **Terminal User Interface** - Beautiful TUI for monitoring and management
- 📝 **Request Logging** - Real-time request log streaming
- 🔌 **WebSocket Support** - Real-time communication with TUI client
- ❓ **Built-in Help** - Comprehensive help system in TUI

## Quick Start

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd webserver

# Build the application  
go build -o bin/webserver ./cmd/webserver
```

### Running the Server

```bash
# Start with default configuration
./bin/webserver

# Start with custom configuration
./bin/webserver -config /path/to/config.json

# Show help
./bin/webserver -help
```

### Running the TUI Client

```bash
# Connect to local server
./bin/webserver --client

# Connect to remote server
./bin/webserver --client -server ws://example.com:8080/ws

# Show help
./bin/webserver -help
```

### Quick Test

```bash
# Run the included test script to see the TUI in action
./test_tui.sh

# Test scrolling functionality
./test_tui_with_scrolling.sh

# Test advanced filtering features
./test_filtering.sh

# Then in another terminal:
./bin/webserver --client
```

## Configuration

The server uses JSON configuration files with the following structure:

```json
{
  "server": {
    "port": 8080,
    "host": "0.0.0.0",
    "static_dir": "./static"
  },
  "endpoints": {
    "/api/error": {
      "type": "error",
      "status_code": 500,
      "message": "Internal Server Error"
    },
    "/api/delay": {
      "type": "delay",
      "delay_ms": 2000,
      "response": {
        "message": "Delayed response"
      }
    },
    "/api/flaky": {
      "type": "conditional_error",
      "error_every_n": 3,
      "status_code": 503,
      "success_response": {
        "status": "ok"
      }
    }
  }
}
```

### Endpoint Types

#### Error Endpoint
Returns a specific HTTP error code:
```json
{
  "type": "error",
  "status_code": 500,
  "message": "Custom error message"
}
```

#### Delay Endpoint
Adds a delay before responding:
```json
{
  "type": "delay",
  "delay_ms": 1000,
  "response": {
    "message": "This response was delayed"
  }
}
```

#### Conditional Error Endpoint
Returns an error every N requests:
```json
{
  "type": "conditional_error",
  "error_every_n": 3,
  "status_code": 503,
  "success_response": {
    "status": "ok"
  }
}
```

## API Endpoints

### Configuration Management

- `GET /config` - Get current configuration
- `PUT /config` - Update entire configuration
- `POST /config` - Add/update a specific endpoint
- `DELETE /config?path=/api/endpoint` - Remove an endpoint

### Statistics and Monitoring

- `GET /stats` - Get server statistics
- `GET /ws` - WebSocket connection for TUI

### Example API Usage

```bash
# Get current configuration
curl http://localhost:8080/config

# Get server statistics
curl http://localhost:8080/stats

# Add a new endpoint
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/test",
    "config": {
      "type": "error",
      "status_code": 404,
      "message": "Test endpoint"
    }
  }'

# Remove an endpoint
curl -X DELETE "http://localhost:8080/config?path=/api/test"
```

## Terminal User Interface

The TUI provides real-time monitoring with multiple tabs:

### Overview Tab
- Server information and uptime
- Quick statistics summary
- Recent activity log

### Configuration Tab
- Current server configuration
- Endpoint configurations with details
- Real-time configuration updates

### Statistics Tab
- Overall server statistics
- Per-endpoint metrics
- Response time analysis
- Status code distribution

### Request Log Tab
- Real-time request streaming (1-second updates)
- Time-ordered display (newest first)
- Advanced text filtering with debouncing
- Toggle to hide /stats endpoint requests
- Color-coded by status code with text highlighting
- Detailed request information with summaries

### Help Tab
- Keyboard shortcuts reference
- Tab descriptions and usage
- Connection information
- Status code color legend
- Troubleshooting guide

### Keyboard Shortcuts

#### Navigation
- `Tab` / `Shift+Tab` - Switch between tabs
- `R` - Refresh data
- `Q` / `Ctrl+C` - Quit

#### Scrolling
- `↑` / `k` - Scroll up one line
- `↓` / `j` - Scroll down one line  
- `Page Up` / `u` - Scroll up half page
- `Page Down` / `d` - Scroll down half page
- `Home` / `g` - Go to top
- `End` / `G` - Go to bottom

#### Request Log Filtering (Request Log tab only)
- `F` - Enter/exit filter mode (type to search)
- `S` - Toggle hide /stats requests
- `C` - Clear all filters
- `Enter` / `Esc` - Exit filter mode
- `Backspace` - Delete filter characters

### TUI Features
- **Real-time Data**: Auto-refreshes every 1 second for faster updates
- **Full Scrolling Support**: Navigate through long content with vim-style keys
- **Advanced Filtering**: Text search with 200ms debouncing and /stats toggle
- **Smart Highlighting**: Matching filter text highlighted in real-time
- **Time Ordering**: Requests automatically sorted by timestamp (newest first)
- **Per-tab Scroll Memory**: Each tab remembers its scroll position
- **Scroll Indicators**: Visual indicators (▲▼) show when more content is available
- **Color Coding**: Status codes are color-coded for easy identification
- **Error Handling**: Graceful connection retry and error display
- **Responsive Design**: Adapts to terminal size with dynamic content layout
- **Built-in Help**: Comprehensive help system with troubleshooting

## Development

### Project Structure

```
webserver/
├── cmd/
│   └── webserver/         # Unified binary (server + client)
├── internal/
│   ├── config/         # Configuration management
│   ├── server/         # HTTP server and handlers
│   └── tui/            # Terminal user interface
├── pkg/
│   └── types/          # Shared types and structures
├── tests/
│   ├── unit/           # Unit tests
│   └── integration/    # Integration tests
├── configs/            # Configuration files
├── static/             # Static file directory
├── test_tui.sh         # Basic TUI testing script
├── test_tui_with_scrolling.sh # Comprehensive scrolling test script
├── test_filtering.sh   # Advanced filtering features test script
└── README.md
```

### Testing

```bash
# Run all tests
go test ./...

# Run unit tests only
go test ./tests/unit/...

# Run integration tests only
go test ./tests/integration/...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Test TUI with real data
./test_tui.sh

# Test scrolling functionality
./test_tui_with_scrolling.sh

# Test advanced filtering features
./test_filtering.sh

# Then in another terminal:
./bin/webserver --client
```

### Building

```bash
# Build unified binary
go build -o bin/webserver ./cmd/webserver

# Build with version info
go build -ldflags "-X main.version=1.0.0" -o bin/webserver ./cmd/webserver

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/webserver-linux ./cmd/webserver
GOOS=windows GOARCH=amd64 go build -o bin/webserver-windows.exe ./cmd/webserver
GOOS=darwin GOARCH=amd64 go build -o bin/webserver-macos ./cmd/webserver
```

## Docker

The application includes full Docker support with production-ready configuration.

### Quick Start with Docker

```bash
# Build and run with Docker Compose (recommended)
docker-compose up -d

# Or build and run with Docker directly
docker build -t webserver .
docker run -d -p 8080:8080 --name webserver-server webserver
```

### Docker Features

- **Multi-stage build**: Optimized image size using golang:1.24-alpine for build and alpine:latest for runtime
- **Security**: Non-root user execution with minimal attack surface
- **Health checks**: Built-in health monitoring via `/stats` endpoint
- **Configuration mounting**: Easy config customization via volume mounts
- **Development mode**: Special dev profile with volume mounts for live development

### Docker Compose Services

```bash
# Production service
docker-compose up -d webserver

# Development service (with live code mounting)
docker-compose --profile dev up -d webserver-dev

# View logs
docker-compose logs -f webserver

# Stop services
docker-compose down
```

### Custom Configuration

```bash
# Run with custom configuration file
docker run -d -p 8080:8080 \
  -v /path/to/custom-config.json:/app/configs/default.json:ro \
  webserver

# Or with Docker Compose (uncomment volume in docker-compose.yml)
# volumes:
#   - ./configs/custom.json:/app/configs/default.json:ro
```

### Testing Docker Setup

```bash
# Run the Docker test script
./test_docker.sh
```

## Use Cases

### Development and Testing
- **API Mocking**: Create mock endpoints that return specific responses
- **Error Simulation**: Test error handling with configurable error rates
- **Performance Testing**: Add delays to simulate slow responses
- **Load Testing**: Monitor request patterns and response times

### DevOps and Monitoring
- **Service Health Checks**: Monitor endpoint availability and performance
- **Configuration Management**: Hot-reload configurations without downtime
- **Request Analysis**: Analyze request patterns and identify issues
- **System Testing**: Test system behavior under different conditions

### Educational and Training
- **HTTP Learning**: Understand HTTP status codes and response patterns
- **Server Monitoring**: Learn about server metrics and performance monitoring
- **Configuration Management**: Practice with configuration-driven applications
- **Real-time Systems**: Explore WebSocket communication and real-time updates

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Write tests for new features
- Follow Go best practices and conventions
- Update documentation for API changes
- Test with multiple configurations
- Ensure backward compatibility

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Go](https://golang.org/)
- TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- WebSocket support via [Gorilla WebSocket](https://github.com/gorilla/websocket)
- File watching with [fsnotify](https://github.com/fsnotify/fsnotify)
- Testing with [Testify](https://github.com/stretchr/testify)

## Version History

- **1.0.0** - Initial release
  - Core server functionality
  - Configuration management
  - Terminal user interface with help system
  - WebSocket support
  - Comprehensive testing suite
  - Real-time data display and monitoring 