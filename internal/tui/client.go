package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"webserver/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the TUI application state
type Model struct {
	// Connection
	serverURL string
	httpURL   string
	connected bool

	// Application state
	config     *types.Config
	stats      *types.ServerStats
	requestLog []types.RequestLogEntry

	// UI state
	activeTab int
	width     int
	height    int

	// Scrolling state
	scrollPositions []int // scroll position for each tab
	contentHeights  []int // content height for each tab
	viewportHeight  int   // available height for content

	// Request log filtering state
	filterMode        bool      // whether we're in filter input mode
	filterText        string    // current filter text
	filterBuffer      string    // typing buffer for debouncing
	hideStatsRequests bool      // toggle to hide /stats requests
	lastFilterUpdate  time.Time // for debouncing

	// Configuration filtering state
	configFilterMode       bool      // whether we're in config filter input mode
	configFilterText       string    // current config filter text
	configFilterBuffer     string    // typing buffer for debouncing
	lastConfigFilterUpdate time.Time // for debouncing

	// Auto-refresh state
	autoRefresh  bool // whether auto-refresh is enabled
	manualScroll bool // whether user has manually scrolled

	// Styles
	tabStyle       lipgloss.Style
	activeTabStyle lipgloss.Style
	contentStyle   lipgloss.Style
	headerStyle    lipgloss.Style
	filterStyle    lipgloss.Style

	// Error state
	lastError string
}

// Tab represents a tab in the TUI
type Tab struct {
	Name string
	View func(*Model) string
}

var tabs = []Tab{
	{"Overview", (*Model).overviewView},
	{"Configuration", (*Model).configView},
	{"Statistics", (*Model).statsView},
	{"Request Log", (*Model).requestLogView},
	{"Help", (*Model).helpView},
}

// NewModel creates a new TUI model
func NewModel(serverURL string) *Model {
	// Convert WebSocket URL to HTTP URL
	httpURL := strings.Replace(serverURL, "ws://", "http://", 1)
	httpURL = strings.Replace(httpURL, "wss://", "https://", 1)
	httpURL = strings.Replace(httpURL, "/ws", "", 1)

	return &Model{
		serverURL:              serverURL,
		httpURL:                httpURL,
		requestLog:             make([]types.RequestLogEntry, 0),
		scrollPositions:        make([]int, len(tabs)),
		contentHeights:         make([]int, len(tabs)),
		viewportHeight:         20, // Default height, will be updated
		filterMode:             false,
		filterText:             "",
		filterBuffer:           "",
		hideStatsRequests:      false,
		lastFilterUpdate:       time.Now(),
		configFilterMode:       false,
		configFilterText:       "",
		configFilterBuffer:     "",
		lastConfigFilterUpdate: time.Now(),
		autoRefresh:            true, // Auto-refresh is enabled by default
		manualScroll:           false,
		tabStyle: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("#3C3C3C")).
			Foreground(lipgloss.Color("#FFFFFF")),
		activeTabStyle: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("#7C7C7C")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true),
		contentStyle: lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C7C7C")),
		headerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5F5F5F")).
			Padding(0, 1).
			Bold(true),
		filterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1).
			Bold(true),
	}
}

// Init initializes the TUI model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.connectToServer,
		tea.EnterAltScreen,
		tea.Tick(time.Second*1, func(time.Time) tea.Msg { return RefreshMsg{} }),               // Update every 1 second
		tea.Tick(time.Millisecond*200, func(time.Time) tea.Msg { return FilterDebounceMsg{} }), // Debounce timer
	)
}

// Update handles TUI updates
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate viewport height (total height - header - status - tabs - footer)
		m.viewportHeight = msg.Height - 8 // Reserve more space for filter UI
		if m.viewportHeight < 5 {
			m.viewportHeight = 5
		}
		return m, nil

	case tea.KeyMsg:
		// Handle filter mode input
		if m.filterMode && m.activeTab == 3 { // Request Log tab
			switch msg.String() {
			case "enter", "esc":
				m.filterMode = false
				m.filterText = m.filterBuffer
				return m, nil
			case "backspace":
				if len(m.filterBuffer) > 0 {
					m.filterBuffer = m.filterBuffer[:len(m.filterBuffer)-1]
					m.lastFilterUpdate = time.Now()
				}
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			default:
				m.filterBuffer += msg.String()
				m.lastFilterUpdate = time.Now()
				return m, nil
			}
		}

		// Handle configuration filter mode input
		if m.configFilterMode && m.activeTab == 1 { // Configuration tab
			switch msg.String() {
			case "enter", "esc":
				m.configFilterMode = false
				m.configFilterText = m.configFilterBuffer
				return m, nil
			case "backspace":
				if len(m.configFilterBuffer) > 0 {
					m.configFilterBuffer = m.configFilterBuffer[:len(m.configFilterBuffer)-1]
					m.lastConfigFilterUpdate = time.Now()
				}
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			default:
				m.configFilterBuffer += msg.String()
				m.lastConfigFilterUpdate = time.Now()
				return m, nil
			}
		}

		// Normal mode key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(tabs)
			return m, nil
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + len(tabs)) % len(tabs)
			return m, nil
		case "up", "k":
			// Scroll up
			if m.scrollPositions[m.activeTab] > 0 {
				m.scrollPositions[m.activeTab]--
				// Disable auto-refresh when user scrolls in Request Log tab
				if m.activeTab == 3 { // Request Log tab
					m.manualScroll = true
					m.autoRefresh = false
				}
			}
			return m, nil
		case "down", "j":
			// Scroll down
			maxScroll := m.contentHeights[m.activeTab] - m.viewportHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollPositions[m.activeTab] < maxScroll {
				m.scrollPositions[m.activeTab]++
				// Disable auto-refresh when user scrolls in Request Log tab
				if m.activeTab == 3 { // Request Log tab
					m.manualScroll = true
					m.autoRefresh = false
				}
			}
			return m, nil
		case "pgup", "u":
			// Page up
			m.scrollPositions[m.activeTab] -= m.viewportHeight / 2
			if m.scrollPositions[m.activeTab] < 0 {
				m.scrollPositions[m.activeTab] = 0
			}
			// Disable auto-refresh when user scrolls in Request Log tab
			if m.activeTab == 3 { // Request Log tab
				m.manualScroll = true
				m.autoRefresh = false
			}
			return m, nil
		case "pgdown", "d":
			// Page down
			maxScroll := m.contentHeights[m.activeTab] - m.viewportHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.scrollPositions[m.activeTab] += m.viewportHeight / 2
			if m.scrollPositions[m.activeTab] > maxScroll {
				m.scrollPositions[m.activeTab] = maxScroll
			}
			// Disable auto-refresh when user scrolls in Request Log tab
			if m.activeTab == 3 { // Request Log tab
				m.manualScroll = true
				m.autoRefresh = false
			}
			return m, nil
		case "home", "g":
			// Go to top
			m.scrollPositions[m.activeTab] = 0
			// Disable auto-refresh when user scrolls in Request Log tab
			if m.activeTab == 3 { // Request Log tab
				m.manualScroll = true
				m.autoRefresh = false
			}
			return m, nil
		case "end", "G":
			// Go to bottom
			maxScroll := m.contentHeights[m.activeTab] - m.viewportHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.scrollPositions[m.activeTab] = maxScroll
			// Disable auto-refresh when user scrolls in Request Log tab
			if m.activeTab == 3 { // Request Log tab
				m.manualScroll = true
				m.autoRefresh = false
			}
			return m, nil
		case "r":
			// Refresh data
			// If we're in the request log tab, also reset the log generation flag to get fresh timestamps
			if m.activeTab == 3 { // Request Log tab
				// No-op, log generation is removed
			}
			return m, tea.Batch(m.fetchConfig, m.fetchStats, m.fetchRequestLog)
		case "a":
			// Toggle auto-refresh (only in Request Log tab)
			if m.activeTab == 3 {
				m.autoRefresh = !m.autoRefresh
				if m.autoRefresh {
					// When re-enabling auto-refresh, reset manual scroll flag
					m.manualScroll = false
				}
			}
			return m, nil
		case "f":
			// Toggle filter mode (Request Log and Configuration tabs)
			if m.activeTab == 3 { // Request Log tab
				m.filterMode = !m.filterMode
				if m.filterMode {
					m.filterBuffer = m.filterText
				}
			} else if m.activeTab == 1 { // Configuration tab
				m.configFilterMode = !m.configFilterMode
				if m.configFilterMode {
					m.configFilterBuffer = m.configFilterText
				}
			}
			return m, nil
		case "s":
			// Toggle stats filter (only in Request Log tab)
			if m.activeTab == 3 {
				m.hideStatsRequests = !m.hideStatsRequests
			}
			return m, nil
		case "c":
			// Clear filters
			if m.activeTab == 3 { // Request Log tab
				m.filterText = ""
				m.filterBuffer = ""
			} else if m.activeTab == 1 { // Configuration tab
				m.configFilterText = ""
				m.configFilterBuffer = ""
			}
			return m, nil
		}

	case ConnectedMsg:
		m.connected = true
		m.lastError = ""
		return m, tea.Batch(m.fetchConfig, m.fetchStats, m.fetchRequestLog)

	case DisconnectedMsg:
		m.connected = false
		m.lastError = "Connection lost"
		return m, tea.Tick(time.Second*5, func(time.Time) tea.Msg { return RetryMsg{} })

	case RetryMsg:
		if !m.connected {
			return m, m.connectToServer
		}
		return m, nil

	case RefreshMsg:
		if m.connected {
			// Always fetch config and stats
			cmds := []tea.Cmd{
				m.fetchConfig,
				m.fetchStats,
			}

			// Only fetch request log if auto-refresh is enabled
			if m.autoRefresh {
				cmds = append(cmds, m.fetchRequestLog)
			}

			// Continue the refresh cycle
			cmds = append(cmds, tea.Tick(time.Second*1, func(time.Time) tea.Msg { return RefreshMsg{} }))

			return m, tea.Batch(cmds...)
		}
		return m, tea.Tick(time.Second*1, func(time.Time) tea.Msg { return RefreshMsg{} })

	case FilterDebounceMsg:
		// Apply filters after debounce period

		// Request log filter debounce
		if time.Since(m.lastFilterUpdate) >= 200*time.Millisecond && m.filterBuffer != m.filterText {
			m.filterText = m.filterBuffer
		}

		// Configuration filter debounce
		if time.Since(m.lastConfigFilterUpdate) >= 200*time.Millisecond && m.configFilterBuffer != m.configFilterText {
			m.configFilterText = m.configFilterBuffer
		}

		return m, tea.Tick(time.Millisecond*200, func(time.Time) tea.Msg { return FilterDebounceMsg{} })

	case ConfigMsg:
		m.config = msg.Config
		return m, nil

	case StatsMsg:
		m.stats = msg.Stats
		return m, nil

	case RequestLogMsg:
		m.requestLog = msg.Entries
		// Sort by timestamp (newest first)
		sort.Slice(m.requestLog, func(i, j int) bool {
			return m.requestLog[i].Timestamp.After(m.requestLog[j].Timestamp)
		})
		// Mark that we have generated sample log data
		// No-op, log generation is removed
		return m, nil

	case ErrorMsg:
		m.lastError = msg.Error
		return m, nil
	}

	return m, nil
}

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.headerStyle.Width(m.width).Render("WebServer Monitor")

	// Connection status
	connectionStatus := "❌ Disconnected"
	if m.connected {
		connectionStatus = "✅ Connected"
	}

	statusLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render(fmt.Sprintf("Server: %s | Status: %s", m.httpURL, connectionStatus))

	// Error display
	errorLine := ""
	if m.lastError != "" {
		errorLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Render(fmt.Sprintf("Error: %s", m.lastError))
	}

	// Tabs
	var tabViews []string
	for i, tab := range tabs {
		style := m.tabStyle
		if i == m.activeTab {
			style = m.activeTabStyle
		}
		tabViews = append(tabViews, style.Render(tab.Name))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)

	// Filter line (Request Log and Configuration tabs)
	var filterLine string
	if m.activeTab == 3 { // Request Log tab
		filterInfo := ""

		if m.filterMode {
			filterInfo = m.filterStyle.Render(fmt.Sprintf("Filter: %s|", m.filterBuffer))
		} else {
			// Show active filter in green right after "F: Filter"
			if m.filterText != "" {
				filterInfo = fmt.Sprintf("F: Filter '%s'", m.filterText)
				filterInfo = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#00FF00")).
					Render(filterInfo)
			}
		}

		// Build controls with checkbox icons
		var controlParts []string

		// Filter control
		if m.filterText == "" && !m.filterMode {
			controlParts = append(controlParts, "F: Filter")
		}

		// Stats toggle with checkbox
		statsCheckbox := "❌"
		if m.hideStatsRequests {
			statsCheckbox = "✅"
		}
		controlParts = append(controlParts, fmt.Sprintf("S: %s Hide /stats", statsCheckbox))

		// Auto-refresh toggle with checkbox
		autoRefreshCheckbox := "❌"
		if m.autoRefresh {
			autoRefreshCheckbox = "✅"
		}
		controlParts = append(controlParts, fmt.Sprintf("A: %s Auto-refresh", autoRefreshCheckbox))

		// Clear control
		controlParts = append(controlParts, "C: Clear")

		controls := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render(strings.Join(controlParts, " | "))

		if filterInfo != "" {
			filterLine = lipgloss.JoinHorizontal(lipgloss.Left, filterInfo, "  ", controls)
		} else {
			filterLine = controls
		}
	} else if m.activeTab == 1 { // Configuration tab
		filterInfo := ""

		if m.configFilterMode {
			filterInfo = m.filterStyle.Render(fmt.Sprintf("Filter: %s|", m.configFilterBuffer))
		} else {
			if m.configFilterText != "" {
				filterInfo = fmt.Sprintf("F: Filter '%s'", m.configFilterText)
				filterInfo = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#00FF00")).
					Render(filterInfo)
			}
		}

		// Build controls
		var controlParts []string

		// Filter control
		if m.configFilterText == "" && !m.configFilterMode {
			controlParts = append(controlParts, "F: Filter")
		}

		// Clear control
		controlParts = append(controlParts, "C: Clear")

		controls := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render(strings.Join(controlParts, " | "))

		if filterInfo != "" {
			filterLine = lipgloss.JoinHorizontal(lipgloss.Left, filterInfo, "  ", controls)
		} else {
			filterLine = controls
		}
	}

	// Content with scrolling
	content := ""
	if m.activeTab < len(tabs) {
		fullContent := tabs[m.activeTab].View(m)
		content = m.renderScrollableContent(fullContent, m.activeTab)
	}

	// Footer with scroll info and filter controls
	footerText := "Tab/Shift+Tab: Switch tabs | ↑↓/j/k: Scroll | PgUp/PgDn/u/d: Page | Home/End/g/G: Top/Bottom | R: Refresh | Q: Quit"
	if m.activeTab == 3 { // Request Log tab
		if m.filterMode {
			footerText = "Filter Mode - Type to filter | Enter/Esc: Exit filter mode | Ctrl+C: Quit"
		} else {
			// Build footer with checkbox status
			statsStatus := "❌"
			if m.hideStatsRequests {
				statsStatus = "✅"
			}
			autoRefreshStatus := "❌"
			if m.autoRefresh {
				autoRefreshStatus = "✅"
			}
			footerText = fmt.Sprintf("F: Filter | S: %s Hide /stats | A: %s Auto-refresh | C: Clear | %s",
				statsStatus, autoRefreshStatus, footerText)
		}
	} else if m.activeTab == 1 { // Configuration tab
		if m.configFilterMode {
			footerText = "Filter Mode - Type to filter endpoints | Enter/Esc: Exit filter mode | Ctrl+C: Quit"
		} else {
			footerText = "F: Filter | C: Clear | " + footerText
		}
	}
	if m.contentHeights[m.activeTab] > m.viewportHeight {
		scrollInfo := fmt.Sprintf(" | Scroll: %d/%d",
			m.scrollPositions[m.activeTab]+1,
			m.contentHeights[m.activeTab]-m.viewportHeight+1)
		footerText += scrollInfo
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render(footerText)

	// Combine all parts
	parts := []string{header, statusLine}
	if errorLine != "" {
		parts = append(parts, errorLine)
	}
	parts = append(parts, tabBar)
	if filterLine != "" {
		parts = append(parts, filterLine)
	}
	parts = append(parts, content, footer)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderScrollableContent renders content with scrolling applied
func (m *Model) renderScrollableContent(content string, tabIndex int) string {
	lines := strings.Split(content, "\n")
	m.contentHeights[tabIndex] = len(lines)

	// If content fits in viewport, no scrolling needed
	if len(lines) <= m.viewportHeight {
		m.scrollPositions[tabIndex] = 0
		return m.contentStyle.Render(content)
	}

	// Apply scrolling
	start := m.scrollPositions[tabIndex]
	end := start + m.viewportHeight

	if start < 0 {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]
	scrolledContent := strings.Join(visibleLines, "\n")

	// Add scroll indicators
	scrollIndicator := ""
	if m.scrollPositions[tabIndex] > 0 {
		scrollIndicator += "▲ "
	}
	if end < len(lines) {
		scrollIndicator += "▼"
	}

	if scrollIndicator != "" {
		scrolledContent += "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render(fmt.Sprintf("    %s", scrollIndicator))
	}

	return m.contentStyle.Render(scrolledContent)
}

// filterRequestLog filters the request log based on current filter settings
func (m *Model) filterRequestLog() []types.RequestLogEntry {
	if len(m.requestLog) == 0 {
		return m.requestLog
	}

	filtered := make([]types.RequestLogEntry, 0)

	for _, entry := range m.requestLog {
		// Skip /stats requests if toggle is enabled
		if m.hideStatsRequests && (strings.Contains(entry.Path, "/stats") || strings.Contains(entry.Path, "/requestlog") || strings.Contains(entry.Path, "/config")) {
			continue
		}

		// Apply text filter if set
		if m.filterText != "" {
			filterLower := strings.ToLower(m.filterText)
			if !strings.Contains(strings.ToLower(entry.Path), filterLower) &&
				!strings.Contains(strings.ToLower(entry.Method), filterLower) &&
				!strings.Contains(strings.ToLower(entry.RemoteAddr), filterLower) {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// filterConfigEndpoints filters configuration endpoints based on current filter settings
func (m *Model) filterConfigEndpoints() map[string]types.EndpointConfig {
	if m.config == nil || len(m.config.Endpoints) == 0 {
		return make(map[string]types.EndpointConfig)
	}

	filtered := make(map[string]types.EndpointConfig)

	for path, endpoint := range m.config.Endpoints {
		// Apply text filter if set
		if m.configFilterText != "" {
			filterLower := strings.ToLower(m.configFilterText)
			if !strings.Contains(strings.ToLower(path), filterLower) &&
				!strings.Contains(strings.ToLower(endpoint.Type), filterLower) &&
				!strings.Contains(strings.ToLower(endpoint.Message), filterLower) {
				continue
			}
		}

		filtered[path] = endpoint
	}

	return filtered
}

// connectToServer connects to the server
func (m *Model) connectToServer() tea.Msg {
	// Test connection by making a simple HTTP request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(m.httpURL + "/stats")
	if err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to connect: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrorMsg{Error: fmt.Sprintf("Server returned status: %d", resp.StatusCode)}
	}

	return ConnectedMsg{}
}

// fetchConfig fetches configuration from the server
func (m *Model) fetchConfig() tea.Msg {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(m.httpURL + "/config")
	if err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to fetch config: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrorMsg{Error: fmt.Sprintf("Config request failed: %d", resp.StatusCode)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to read config response: %v", err)}
	}

	var config types.Config
	if err := json.Unmarshal(body, &config); err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to parse config: %v", err)}
	}

	return ConfigMsg{Config: &config}
}

// fetchStats fetches statistics from the server
func (m *Model) fetchStats() tea.Msg {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(m.httpURL + "/stats")
	if err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to fetch stats: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrorMsg{Error: fmt.Sprintf("Stats request failed: %d", resp.StatusCode)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to read stats response: %v", err)}
	}

	var stats types.ServerStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to parse stats: %v", err)}
	}

	return StatsMsg{Stats: &stats}
}

// fetchRequestLog fetches real request log data from the server
func (m *Model) fetchRequestLog() tea.Msg {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(m.httpURL + "/requestlog")
	if err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to fetch request log: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrorMsg{Error: fmt.Sprintf("Request log request failed: %d", resp.StatusCode)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to read request log response: %v", err)}
	}

	var requestLog []types.RequestLogEntry
	if err := json.Unmarshal(body, &requestLog); err != nil {
		return ErrorMsg{Error: fmt.Sprintf("Failed to parse request log: %v", err)}
	}

	return RequestLogMsg{Entries: requestLog}
}

// Helper function
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// Message types for TUI communication
type ConnectedMsg struct{}
type DisconnectedMsg struct{}
type RetryMsg struct{}
type RefreshMsg struct{}
type FilterDebounceMsg struct{}
type ConfigMsg struct{ Config *types.Config }
type StatsMsg struct{ Stats *types.ServerStats }
type RequestLogMsg struct{ Entries []types.RequestLogEntry }
type ErrorMsg struct{ Error string }

// RunTUI starts the TUI application
func RunTUI(serverURL string) error {
	model := NewModel(serverURL)

	p := tea.NewProgram(model, tea.WithAltScreen())

	// Start the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	return nil
}
