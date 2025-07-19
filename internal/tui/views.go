package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// overviewView renders the overview tab
func (m *Model) overviewView() string {
	if !m.connected {
		return "❌ Not connected to server\n\nTry pressing 'R' to refresh or check if the server is running."
	}

	var sections []string

	// Server info
	serverInfo := "📊 Server Overview\n\n"
	if m.config != nil {
		serverInfo += fmt.Sprintf("• Host: %s\n", m.config.Server.Host)
		serverInfo += fmt.Sprintf("• Port: %d\n", m.config.Server.Port)
		serverInfo += fmt.Sprintf("• Static Directory: %s\n", m.config.Server.StaticDir)
		serverInfo += fmt.Sprintf("• Configured Endpoints: %d\n", len(m.config.Endpoints))

		// Add endpoint details
		if len(m.config.Endpoints) > 0 {
			serverInfo += "\n🎯 Endpoint Summary:\n"

			// Sort endpoint paths alphabetically for consistent display
			paths := make([]string, 0, len(m.config.Endpoints))
			for path := range m.config.Endpoints {
				paths = append(paths, path)
			}
			sort.Strings(paths)

			for _, path := range paths {
				endpoint := m.config.Endpoints[path]
				serverInfo += fmt.Sprintf("  • %s (%s)\n", path, endpoint.Type)
			}
		}
	} else {
		serverInfo += "• Loading configuration...\n"
	}

	sections = append(sections, serverInfo)

	// Quick stats
	if m.stats != nil {
		uptime := time.Since(m.stats.StartTime).Truncate(time.Second)
		quickStats := "📈 Quick Statistics\n\n"
		quickStats += fmt.Sprintf("• Uptime: %s\n", uptime)
		quickStats += fmt.Sprintf("• Total Requests: %d\n", m.stats.RequestCount)
		quickStats += fmt.Sprintf("• Total Errors: %d\n", m.stats.ErrorCount)
		if m.stats.RequestCount > 0 {
			errorRate := float64(m.stats.ErrorCount) / float64(m.stats.RequestCount) * 100
			quickStats += fmt.Sprintf("• Error Rate: %.2f%%\n", errorRate)
			avgReqPerMin := float64(m.stats.RequestCount) / (time.Since(m.stats.StartTime).Minutes() + 0.01)
			quickStats += fmt.Sprintf("• Avg Requests/min: %.1f\n", avgReqPerMin)
		}
		quickStats += fmt.Sprintf("• Active Endpoints: %d\n", len(m.stats.Endpoints))

		sections = append(sections, quickStats)
	} else {
		sections = append(sections, "📈 Quick Statistics\n\n• Loading statistics...\n")
	}

	// Recent activity
	recentActivity := "🔄 Recent Activity\n\n"
	if len(m.requestLog) > 0 {
		recentActivity += "Last 5 requests:\n"
		for i, entry := range m.requestLog {
			if i >= 5 { // Show only last 5 entries
				break
			}
			statusEmoji := "✅"
			if entry.StatusCode >= 400 {
				statusEmoji = "❌"
			}
			recentActivity += fmt.Sprintf("%s %s %s (%d) - %dms - %s\n",
				statusEmoji,
				entry.Method,
				entry.Path,
				entry.StatusCode,
				entry.Duration,
				entry.Timestamp.Format("15:04:05"))
		}
	} else {
		recentActivity += "No recent activity\n"
		recentActivity += "\nTo generate activity:\n"
		recentActivity += "• Visit http://localhost:8080/ (static files)\n"
		recentActivity += "• Try http://localhost:8080/api/error (error endpoint)\n"
		recentActivity += "• Try http://localhost:8080/api/delay (delay endpoint)\n"
		recentActivity += "• Try http://localhost:8080/api/flaky (conditional error)\n"
	}

	sections = append(sections, recentActivity)

	// Connection info
	connectionInfo := "🔗 Connection Information\n\n"
	connectionInfo += fmt.Sprintf("• Server URL: %s\n", m.httpURL)
	connectionInfo += fmt.Sprintf("• WebSocket URL: %s\n", m.serverURL)
	connectionInfo += "• Protocol: HTTP polling (every 1 second)\n"
	connectionInfo += "• Connection Status: "
	if m.connected {
		connectionInfo += "✅ Connected\n"
	} else {
		connectionInfo += "❌ Disconnected\n"
	}

	sections = append(sections, connectionInfo)

	content := strings.Join(sections, "\n")
	return content
}

// configView renders the configuration tab
func (m *Model) configView() string {
	if !m.connected {
		return "❌ Not connected to server\n\nTry pressing 'R' to refresh or check if the server is running."
	}

	if m.config == nil {
		return "⏳ Loading configuration..."
	}

	var sections []string

	// Server configuration
	serverConfig := "🔧 Server Configuration\n\n"
	serverConfig += fmt.Sprintf("Host: %s\n", m.config.Server.Host)
	serverConfig += fmt.Sprintf("Port: %d\n", m.config.Server.Port)
	serverConfig += fmt.Sprintf("Static Directory: %s\n", m.config.Server.StaticDir)
	serverConfig += fmt.Sprintf("Full Address: http://%s:%d\n", m.config.Server.Host, m.config.Server.Port)

	sections = append(sections, serverConfig)

	// Endpoints configuration
	endpointsConfig := "🎯 Configured Endpoints\n\n"

	// Get filtered endpoints
	filteredEndpoints := m.filterConfigEndpoints()

	// Show filter status if active
	if m.configFilterText != "" {
		endpointsConfig += fmt.Sprintf("🔍 Filter: '%s' | Showing %d/%d endpoints\n\n",
			m.configFilterText, len(filteredEndpoints), len(m.config.Endpoints))
	}

	if len(m.config.Endpoints) == 0 {
		endpointsConfig += "No endpoints configured\n"
		endpointsConfig += "\nYou can add endpoints using the configuration API:\n"
		endpointsConfig += "curl -X POST http://localhost:8080/config -H 'Content-Type: application/json' \\\n"
		endpointsConfig += "  -d '{\"path\": \"/api/test\", \"config\": {\"type\": \"error\", \"status_code\": 404}}'\n"
	} else if len(filteredEndpoints) == 0 && m.configFilterText != "" {
		endpointsConfig += "🔍 No matching endpoints found\n\n"
		endpointsConfig += fmt.Sprintf("Total endpoints: %d\n", len(m.config.Endpoints))
		endpointsConfig += fmt.Sprintf("Filter: '%s'\n", m.configFilterText)
		endpointsConfig += "\n💡 Tips:\n"
		endpointsConfig += "• Press 'C' to clear filter\n"
		endpointsConfig += "• Press 'F' to change filter\n"
		endpointsConfig += "• Filter matches endpoint path, type, and message\n"
	} else {
		// Sort endpoint paths alphabetically for consistent display
		paths := make([]string, 0, len(filteredEndpoints))
		for path := range filteredEndpoints {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		for _, path := range paths {
			endpoint := filteredEndpoints[path]
			endpointsConfig += fmt.Sprintf("• %s\n", path)
			endpointsConfig += fmt.Sprintf("  Type: %s\n", endpoint.Type)

			switch endpoint.Type {
			case "error":
				endpointsConfig += fmt.Sprintf("  Status Code: %d\n", endpoint.StatusCode)
				if endpoint.Message != "" {
					endpointsConfig += fmt.Sprintf("  Message: %s\n", endpoint.Message)
				}
				endpointsConfig += fmt.Sprintf("  Test: curl http://localhost:8080%s\n", path)
			case "delay":
				endpointsConfig += fmt.Sprintf("  Delay: %dms\n", endpoint.DelayMs)
				if endpoint.Response != nil {
					endpointsConfig += "  Returns: Custom JSON response\n"
				}
				endpointsConfig += fmt.Sprintf("  Test: curl http://localhost:8080%s\n", path)
			case "conditional_error":
				endpointsConfig += fmt.Sprintf("  Error Every N: %d\n", endpoint.ErrorEveryN)
				endpointsConfig += fmt.Sprintf("  Error Status Code: %d\n", endpoint.StatusCode)
				if endpoint.SuccessResponse != nil {
					endpointsConfig += "  Success Response: Custom JSON\n"
				}
				endpointsConfig += fmt.Sprintf("  Test: curl http://localhost:8080%s (multiple times)\n", path)
			}
			endpointsConfig += "\n"
		}

		// Configuration management info
		endpointsConfig += "📝 Configuration Management:\n\n"
		endpointsConfig += "• GET /config - View current configuration\n"
		endpointsConfig += "• PUT /config - Update entire configuration\n"
		endpointsConfig += "• POST /config - Add/update specific endpoint\n"
		endpointsConfig += "• DELETE /config?path=<path> - Remove endpoint\n"
		endpointsConfig += "\nConfiguration is automatically saved to disk and hot-reloaded.\n"
	}

	sections = append(sections, endpointsConfig)

	content := strings.Join(sections, "\n")
	return content
}

// statsView renders the statistics tab
func (m *Model) statsView() string {
	if !m.connected {
		return "❌ Not connected to server\n\nTry pressing 'R' to refresh or check if the server is running."
	}

	if m.stats == nil {
		return "⏳ Loading statistics..."
	}

	var sections []string

	// Overall statistics
	uptime := time.Since(m.stats.StartTime).Truncate(time.Second)
	overallStats := "📊 Overall Statistics\n\n"
	overallStats += fmt.Sprintf("Server Start Time: %s\n", m.stats.StartTime.Format("2006-01-02 15:04:05"))
	overallStats += fmt.Sprintf("Uptime: %s\n", uptime)
	overallStats += fmt.Sprintf("Total Requests: %d\n", m.stats.RequestCount)
	overallStats += fmt.Sprintf("Total Errors: %d\n", m.stats.ErrorCount)
	overallStats += fmt.Sprintf("Success Requests: %d\n", m.stats.RequestCount-m.stats.ErrorCount)

	if m.stats.RequestCount > 0 {
		errorRate := float64(m.stats.ErrorCount) / float64(m.stats.RequestCount) * 100
		successRate := 100.0 - errorRate
		overallStats += fmt.Sprintf("Success Rate: %.2f%%\n", successRate)
		overallStats += fmt.Sprintf("Error Rate: %.2f%%\n", errorRate)

		avgReqPerMin := float64(m.stats.RequestCount) / (time.Since(m.stats.StartTime).Minutes() + 0.01)
		avgReqPerHour := avgReqPerMin * 60
		overallStats += fmt.Sprintf("Avg Requests/min: %.1f\n", avgReqPerMin)
		overallStats += fmt.Sprintf("Avg Requests/hour: %.0f\n", avgReqPerHour)
	}

	sections = append(sections, overallStats)

	// Per-endpoint statistics
	endpointStats := "🎯 Per-Endpoint Statistics\n\n"
	if len(m.stats.Endpoints) == 0 {
		endpointStats += "No endpoint statistics available\n"
		endpointStats += "\nMake some requests to see statistics:\n"
		endpointStats += "• curl http://localhost:8080/api/error\n"
		endpointStats += "• curl http://localhost:8080/api/delay\n"
		endpointStats += "• curl http://localhost:8080/api/flaky\n"
		endpointStats += "• curl http://localhost:8080/\n"
	} else {
		// Sort endpoint paths alphabetically for consistent display
		paths := make([]string, 0, len(m.stats.Endpoints))
		for path := range m.stats.Endpoints {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		for _, path := range paths {
			stats := m.stats.Endpoints[path]
			endpointStats += fmt.Sprintf("━━━ %s ━━━\n", path)
			endpointStats += fmt.Sprintf("Requests: %d\n", stats.RequestCount)
			endpointStats += fmt.Sprintf("Errors: %d\n", stats.ErrorCount)
			endpointStats += fmt.Sprintf("Success: %d\n", stats.RequestCount-stats.ErrorCount)

			if stats.RequestCount > 0 {
				// Response times
				avgTime := float64(stats.TotalTimeMs) / float64(stats.RequestCount)
				endpointStats += fmt.Sprintf("Response Times:\n")
				endpointStats += fmt.Sprintf("  • Average: %.2fms\n", avgTime)
				endpointStats += fmt.Sprintf("  • Minimum: %dms\n", stats.MinTimeMs)
				endpointStats += fmt.Sprintf("  • Maximum: %dms\n", stats.MaxTimeMs)

				// Error rate for this endpoint
				errorRate := float64(stats.ErrorCount) / float64(stats.RequestCount) * 100
				successRate := 100.0 - errorRate
				endpointStats += fmt.Sprintf("Success Rate: %.2f%%\n", successRate)
				endpointStats += fmt.Sprintf("Error Rate: %.2f%%\n", errorRate)

				// Request frequency
				if !stats.FirstRequest.IsZero() && !stats.LastRequest.IsZero() {
					duration := stats.LastRequest.Sub(stats.FirstRequest).Minutes()
					if duration > 0 {
						reqPerMin := float64(stats.RequestCount) / duration
						endpointStats += fmt.Sprintf("Avg Req/min: %.1f\n", reqPerMin)
					}
				}
			}

			// Status code distribution
			if len(stats.StatusCodes) > 0 {
				endpointStats += "Status Code Distribution:\n"

				// Sort status codes for consistent display
				statusCodes := make([]int, 0, len(stats.StatusCodes))
				for code := range stats.StatusCodes {
					statusCodes = append(statusCodes, code)
				}
				sort.Ints(statusCodes)

				for _, code := range statusCodes {
					count := stats.StatusCodes[code]
					percentage := float64(count) / float64(stats.RequestCount) * 100
					endpointStats += fmt.Sprintf("  • %d: %d (%.1f%%)\n", code, count, percentage)
				}
			}

			// Timing information
			if !stats.FirstRequest.IsZero() {
				endpointStats += fmt.Sprintf("First Request: %s\n", stats.FirstRequest.Format("15:04:05"))
			}
			if !stats.LastRequest.IsZero() {
				endpointStats += fmt.Sprintf("Last Request: %s\n", stats.LastRequest.Format("15:04:05"))
			}

			// Conditional error tracking
			if stats.ConditionalCount > 0 {
				endpointStats += fmt.Sprintf("Conditional Counter: %d\n", stats.ConditionalCount)
			}

			endpointStats += "\n"
		}
	}

	sections = append(sections, endpointStats)

	content := strings.Join(sections, "\n")
	return content
}

// requestLogView renders the request log tab
func (m *Model) requestLogView() string {
	if !m.connected {
		return "❌ Not connected to server\n\nTry pressing 'R' to refresh or check if the server is running."
	}

	content := ""

	// Get filtered entries
	filteredEntries := m.filterRequestLog()

	if len(m.requestLog) == 0 {
		content += "No requests logged yet\n\n"
		content += "💡 To generate request log entries:\n"
		content += "• Make requests to the server endpoints\n"
		content += "• Try: curl http://localhost:8080/api/error\n"
		content += "• Try: curl http://localhost:8080/api/delay\n"
		content += "• Try: curl http://localhost:8080/api/flaky\n"
		content += "• Try: curl http://localhost:8080/\n"
		content += "\nThe log will show recent requests with color-coded status codes.\n"
		content += "\n🎨 Status Code Colors:\n"
		content += "• 2xx (Success) - Cyan\n"
		content += "• 3xx (Redirect) - Green\n"
		content += "• 4xx (Client Error) - Yellow\n"
		content += "• 5xx (Server Error) - Red\n"
		content += "\n📋 Filter Controls:\n"
		content += "• F - Enter filter mode (type to search)\n"
		content += "• S - Toggle hide /stats requests\n"
		content += "• C - Clear all filters\n"
	} else if len(filteredEntries) == 0 && (m.filterText != "" || m.hideStatsRequests) {
		content += "🔍 No matching requests found\n\n"
		content += fmt.Sprintf("Total requests: %d\n", len(m.requestLog))
		if m.filterText != "" {
			content += fmt.Sprintf("Filter: '%s'\n", m.filterText)
		}
		if m.hideStatsRequests {
			content += "Hiding /stats requests\n"
		}
		content += "\n💡 Tips:\n"
		content += "• Press 'C' to clear filters\n"
		content += "• Press 'S' to toggle internal endpoints filter\n"
		content += "• Press 'F' to change text filter\n"
		content += "• Press 'A' to toggle auto-refresh on/off\n"
		content += "• Filters match path, method, or IP address\n"
		content += "• Scrolling disables auto-refresh automatically\n"
	} else {
		// Show filter status
		if m.filterText != "" || m.hideStatsRequests {
			statusParts := []string{}
			if m.filterText != "" {
				statusParts = append(statusParts, fmt.Sprintf("Filter: '%s'", m.filterText))
			}
			if m.hideStatsRequests {
				statusParts = append(statusParts, "Hiding /stats")
			}
			content += fmt.Sprintf("🔍 Filtered: %s | Showing %d/%d requests\n\n",
				strings.Join(statusParts, ", "), len(filteredEntries), len(m.requestLog))
		} else {
			content += fmt.Sprintf("📅 Showing all %d requests (⏰ ordered by request time, newest first)\n\n", len(filteredEntries))
		}

		// Header
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5F5F5F")).
			Padding(0, 1)

		header := fmt.Sprintf("%-10s %-8s %-6s %-40s %-6s %-8s %-15s",
			"Time", "Date", "Method", "Path", "Status", "Duration", "Remote")
		content += headerStyle.Render(header) + "\n"

		// Separator line
		content += strings.Repeat("─", 95) + "\n"

		// Log entries (filtered and sorted)
		for i, entry := range filteredEntries {
			timestamp := entry.Timestamp.Format("15:04:05")
			date := entry.Timestamp.Format("01-02")

			// Color code based on status
			var statusColor lipgloss.Color
			switch {
			case entry.StatusCode >= 500:
				statusColor = lipgloss.Color("#FF6B6B") // Red
			case entry.StatusCode >= 400:
				statusColor = lipgloss.Color("#FFD93D") // Yellow
			case entry.StatusCode >= 300:
				statusColor = lipgloss.Color("#6BCF7F") // Green
			default:
				statusColor = lipgloss.Color("#4ECDC4") // Cyan
			}

			statusStyle := lipgloss.NewStyle().Foreground(statusColor).Bold(true)

			// Truncate first, THEN highlight to avoid text disappearing
			truncatedPath := truncateString(entry.Path, 40) // Increased from 25 to 40
			truncatedRemote := truncateString(entry.RemoteAddr, 15)

			// Now apply highlighting to the truncated text
			displayPath := truncatedPath
			displayMethod := entry.Method
			displayRemote := truncatedRemote

			if m.filterText != "" {
				filterLower := strings.ToLower(m.filterText)
				if strings.Contains(strings.ToLower(entry.Path), filterLower) {
					displayPath = highlightText(truncatedPath, m.filterText)
				}
				if strings.Contains(strings.ToLower(entry.Method), filterLower) {
					displayMethod = highlightText(entry.Method, m.filterText)
				}
				if strings.Contains(strings.ToLower(entry.RemoteAddr), filterLower) {
					displayRemote = highlightText(truncatedRemote, m.filterText)
				}
			}

			logLine := fmt.Sprintf("%-10s %-8s %-6s %-40s %-6s %-8s %-15s",
				timestamp,
				date,
				displayMethod,
				displayPath,
				statusStyle.Render(fmt.Sprintf("%d", entry.StatusCode)),
				fmt.Sprintf("%dms", entry.Duration),
				displayRemote)

			content += logLine + "\n"

			// Add separator every 5 entries for readability
			if i > 0 && (i+1)%5 == 0 && i < len(filteredEntries)-1 {
				content += lipgloss.NewStyle().
					Foreground(lipgloss.Color("#666666")).
					Render(strings.Repeat("·", 95)) + "\n" // Updated from 80 to 95
			}
		}

		content += "\n📊 Log Summary:\n"
		if m.filterText != "" || m.hideStatsRequests {
			content += fmt.Sprintf("Filtered Entries: %d (of %d total)\n", len(filteredEntries), len(m.requestLog))
		} else {
			content += fmt.Sprintf("Total Entries: %d\n", len(filteredEntries))
		}

		// Count status codes in filtered results
		statusCounts := make(map[int]int)
		methodCounts := make(map[string]int)
		for _, entry := range filteredEntries {
			statusCounts[entry.StatusCode]++
			methodCounts[entry.Method]++
		}

		// Sort status codes and methods
		statusCodes := make([]int, 0, len(statusCounts))
		for code := range statusCounts {
			statusCodes = append(statusCodes, code)
		}
		sort.Ints(statusCodes)

		methods := make([]string, 0, len(methodCounts))
		for method := range methodCounts {
			methods = append(methods, method)
		}
		sort.Strings(methods)

		if len(statusCodes) > 0 {
			content += "Status Code Distribution:\n"
			for _, code := range statusCodes {
				percentage := float64(statusCounts[code]) / float64(len(filteredEntries)) * 100
				content += fmt.Sprintf("  • %d: %d entries (%.1f%%)\n", code, statusCounts[code], percentage)
			}
		}

		if len(methods) > 1 {
			content += "HTTP Methods:\n"
			for _, method := range methods {
				percentage := float64(methodCounts[method]) / float64(len(filteredEntries)) * 100
				content += fmt.Sprintf("  • %s: %d (%.1f%%)\n", method, methodCounts[method], percentage)
			}
		}

		// Time range info
		if len(filteredEntries) > 0 {
			oldest := filteredEntries[len(filteredEntries)-1].Timestamp
			newest := filteredEntries[0].Timestamp
			timeSpan := newest.Sub(oldest)
			content += fmt.Sprintf("Time Range: %s (spanning %s)\n",
				oldest.Format("15:04:05"),
				timeSpan.Truncate(time.Second))
		}
	}

	return content
}

// highlightText highlights matching text in the original string
func highlightText(original, filter string) string {
	if filter == "" || original == "" {
		return original
	}

	// Case-insensitive search
	filterLower := strings.ToLower(filter)
	originalLower := strings.ToLower(original)

	if !strings.Contains(originalLower, filterLower) {
		return original
	}

	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FFFF00")).
		Foreground(lipgloss.Color("#000000")).
		Bold(true)

	// Find the actual case-preserved match
	index := strings.Index(originalLower, filterLower)
	if index >= 0 && index+len(filter) <= len(original) {
		before := original[:index]
		match := original[index : index+len(filter)]
		after := original[index+len(filter):]
		return before + highlightStyle.Render(match) + after
	}

	return original
}

// helpView renders the help tab
func (m *Model) helpView() string {
	content := "❓ Help & Controls\n\n"

	// Keyboard shortcuts
	content += "⌨️  Keyboard Shortcuts:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += "Navigation:\n"
	content += "• Tab             - Switch to next tab\n"
	content += "• Shift+Tab       - Switch to previous tab\n"
	content += "\nScrolling:\n"
	content += "• ↑ / k           - Scroll up one line\n"
	content += "• ↓ / j           - Scroll down one line\n"
	content += "• Page Up / u     - Scroll up half page\n"
	content += "• Page Down / d   - Scroll down half page\n"
	content += "• Home / g        - Go to top\n"
	content += "• End / G         - Go to bottom\n"
	content += "\nFiltering:\n"
	content += "• F               - Enter/exit filter mode (Request Log & Configuration tabs)\n"
	content += "• C               - Clear all filters (Request Log & Configuration tabs)\n"
	content += "• Enter/Esc       - Exit filter mode (in filter mode)\n"
	content += "• Backspace       - Delete filter characters (in filter mode)\n"
	content += "\nRequest Log Specific:\n"
	content += "• S               - Toggle hide /stats requests\n"
	content += "• A               - Toggle auto-refresh on/off\n"
	content += "\nActions:\n"
	content += "• R               - Refresh data from server\n"
	content += "• Q / Ctrl+C      - Quit application\n\n"

	// Tab descriptions
	content += "📑 Tab Descriptions:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += "• Overview        - Server info, quick stats, recent activity\n"
	content += "                    Shows server configuration, uptime, request counts,\n"
	content += "                    and the last few requests made to the server.\n\n"
	content += "• Configuration   - Server settings and endpoint configurations\n"
	content += "                    View current server config and all configured\n"
	content += "                    dynamic endpoints with their settings.\n\n"
	content += "• Statistics      - Detailed per-endpoint metrics and performance\n"
	content += "                    Comprehensive statistics including response times,\n"
	content += "                    error rates, and request frequency per endpoint.\n\n"
	content += "• Request Log     - Real-time request log with advanced filtering\n"
	content += "                    Shows recent HTTP requests with timestamps,\n"
	content += "                    methods, paths, status codes, and durations.\n"
	content += "                    Auto-updates every 1 second. Supports text filtering\n"
	content += "                    and toggling /stats requests visibility.\n\n"
	content += "• Help            - This help screen with shortcuts and info\n"
	content += "                    Complete reference for using the TUI.\n\n"

	// Request log filtering section
	content += "🔍 Filtering Capabilities:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += "Both Request Log and Configuration tabs support text filtering:\n\n"
	content += "Text Filtering (Both tabs):\n"
	content += "• Press 'F' to enter filter mode\n"
	content += "• Type to search through relevant fields\n"
	content += "• Filter applies automatically with 200ms debouncing\n"
	content += "• Matching text is highlighted in yellow\n"
	content += "• Press Enter or Esc to exit filter mode\n"
	content += "• Press 'C' to clear filters\n\n"
	content += "Request Log Filtering:\n"
	content += "• Filters: paths, methods, and IP addresses\n"
	content += "• Additional 'S' key to hide/show /stats endpoints\n"
	content += "• Auto-refresh toggle with 'A' key\n"
	content += "• Status shown: 'Showing X/Y requests'\n\n"
	content += "Configuration Filtering:\n"
	content += "• Filters: endpoint paths, types, and messages\n"
	content += "• Useful for finding specific endpoints in large configurations\n"
	content += "• Status shown: 'Showing X/Y endpoints'\n"
	content += "• Maintains alphabetical sorting of filtered results\n\n"
	content += "Auto-Refresh Toggle (Request Log only):\n"
	content += "• Press 'A' to toggle auto-refresh on/off\n"
	content += "• When ON: New requests automatically appear every 1 second\n"
	content += "• When OFF: Manual refresh required (press 'R')\n"
	content += "• Auto-refresh disables automatically when you scroll\n"
	content += "• Status shown in header: 'auto-refresh: ON/OFF'\n\n"
	content += "Clear Filters:\n"
	content += "• Press 'C' to clear all active filters\n\n"
	content += "Filter Indicators:\n"
	content += "• Active filters shown below tabs in green\n"
	content += "• Filter mode shown in yellow with typing cursor\n"
	content += "• Filtered count displayed: 'Showing X/Y requests'\n\n"

	// Connection info
	content += "🔗 Connection Information:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += fmt.Sprintf("• Server URL:     %s\n", m.httpURL)
	content += fmt.Sprintf("• WebSocket URL:  %s\n", m.serverURL)
	content += "• Protocol:       HTTP polling (every 1 second)\n"
	content += "• Status:         "
	if m.connected {
		content += "✅ Connected\n"
	} else {
		content += "❌ Disconnected\n"
	}
	content += "• Auto-refresh:   Every 1 second\n\n"

	// Status indicators
	content += "🎨 Status Code Colors:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += "• 2xx Success     - " + lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Render("Cyan") + "   (200 OK, 201 Created, etc.)\n"
	content += "• 3xx Redirect    - " + lipgloss.NewStyle().Foreground(lipgloss.Color("#6BCF7F")).Render("Green") + "  (301 Moved, 302 Found, etc.)\n"
	content += "• 4xx Client Err  - " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Render("Yellow") + " (400 Bad Request, 404 Not Found, etc.)\n"
	content += "• 5xx Server Err  - " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Red") + "    (500 Internal Error, 503 Unavailable, etc.)\n\n"

	// API endpoints
	content += "🌐 Server API Endpoints:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += "• GET /config     - Get current server configuration\n"
	content += "• PUT /config     - Update entire configuration\n"
	content += "• POST /config    - Add/update specific endpoint\n"
	content += "• DELETE /config  - Remove endpoint (?path=<path>)\n"
	content += "• GET /stats      - Get server statistics\n"
	content += "• GET /ws         - WebSocket connection (for future real-time updates)\n\n"

	// Troubleshooting
	content += "🔧 Troubleshooting:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += "• Not Connected?  - Check if server is running on the specified URL\n"
	content += "                    Try: ./webserver (in another terminal)\n"
	content += "• No Data?        - Press 'R' to refresh or wait for auto-refresh\n"
	content += "• Slow Updates?   - Network latency may cause delays\n"
	content += "• TUI Issues?     - Try resizing terminal window\n"
	content += "• Text Cut Off?   - Use scroll keys (↑↓) or resize terminal\n"
	content += "• Log Empty?      - Make requests to server endpoints to see logs\n\n"

	// Tips
	content += "💡 Pro Tips:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += "• Use vim-style keys (j/k) for comfortable scrolling\n"
	content += "• Check Overview tab for quick server health summary\n"
	content += "• Statistics tab shows detailed performance metrics\n"
	content += "• Request Log updates every 1 second with real-time data\n"
	content += "• Use Request Log filtering to focus on specific endpoints\n"
	content += "• Toggle internal endpoints visibility to see only application requests\n"
	content += "• All data auto-refreshes - no need to manually refresh often\n"
	content += "• Scroll position is remembered when switching between tabs\n\n"

	// About
	content += "ℹ️  About WebServer:\n"
	content += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	content += "WebServer Configurable Web Server TUI Client\n"
	content += "Version: 1.0.0\n"
	content += "Built with Go, Bubble Tea, and WebSocket\n"
	content += "\nFeatures:\n"
	content += "• Configurable static file serving\n"
	content += "• Dynamic response generation (errors, delays, conditional)\n"
	content += "• Hot configuration reloading\n"
	content += "• Real-time statistics and monitoring\n"
	content += "• Beautiful terminal user interface\n"
	content += "• Comprehensive help system (this screen!)\n"

	return content
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
