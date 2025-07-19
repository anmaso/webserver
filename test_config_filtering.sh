#!/bin/bash

echo "🔍 Testing Configuration Tab Filtering..."
echo ""
echo "✅ New Feature Added:"
echo "  • Text filtering for Configuration tab endpoints"
echo "  • F key to enter/exit filter mode"
echo "  • C key to clear filters"
echo "  • Real-time filtering with 200ms debouncing"
echo "  • Visual filter indicators and status"
echo ""

# Start the server
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Creating diverse endpoint configuration for filtering tests..."

# Add various endpoints with different types, paths, and messages to test filtering
curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/user/login", "config": {"type": "error", "status_code": 401, "message": "Authentication required"}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/user/profile", "config": {"type": "delay", "delay_ms": 150}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/admin/dashboard", "config": {"type": "conditional_error", "error_every_n": 4, "status_code": 403}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/data/export", "config": {"type": "delay", "delay_ms": 2000}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/health", "config": {"type": "error", "status_code": 200, "message": "Service healthy"}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/metrics", "config": {"type": "delay", "delay_ms": 50}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/payment/process", "config": {"type": "conditional_error", "error_every_n": 2, "status_code": 500}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/notification/send", "config": {"type": "error", "status_code": 429, "message": "Rate limit exceeded"}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/file/upload", "config": {"type": "delay", "delay_ms": 1000}}' > /dev/null

sleep 2

echo ""
echo "🎯 Test Configuration Filtering:"
echo ""
echo "┌─ Basic Filtering ─────────────────────────────────────────────────┐"
echo "│ 1. Start TUI: ./bin/webserver --client                               │"
echo "│ 2. Go to Configuration tab (should show 9 endpoints)              │"
echo "│ 3. Press 'F' to enter filter mode                                 │"
echo "│ 4. Type 'api' - should show only /api/* endpoints                 │"
echo "│ 5. Press Esc to exit filter mode                                  │"
echo "│ 6. Filter should remain active, showing filtered count            │"
echo "└────────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Different Filter Tests ──────────────────────────────────────────┐"
echo "│ Test these filters (press F, type, then Esc):                     │"
echo "│ • 'user' - Shows: /api/user/login, /api/user/profile             │"
echo "│ • 'delay' - Shows all delay type endpoints                        │"
echo "│ • 'error' - Shows all error type endpoints                        │"
echo "│ • 'admin' - Shows: /api/admin/dashboard                          │"
echo "│ • 'Auth' - Shows: /api/user/login (matches message)              │"
echo "│ • 'health' - Shows: /health                                       │"
echo "│ • '429' - Shows: /api/notification/send (matches status)         │"
echo "└────────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Filter Management ───────────────────────────────────────────────┐"
echo "│ • Press 'C' to clear any active filter                            │"
echo "│ • Press 'F' again to change filter text                          │"
echo "│ • Use backspace in filter mode to delete characters              │"
echo "│ • Filter status shows below tabs in green when active            │"
echo "│ • Yellow filter mode indicator when typing                        │"
echo "└────────────────────────────────────────────────────────────────────┘"
echo ""
echo "🔧 Features to Verify:"
echo ""
echo "✅ Filter Matching:"
echo "   • Searches endpoint paths (/api/user/login)"
echo "   • Searches endpoint types (error, delay, conditional_error)"
echo "   • Searches error messages (Authentication required)"
echo "   • Case-insensitive matching"
echo ""
echo "✅ Visual Indicators:"
echo "   • Green 'Active: Filter: ...' when filter is set"
echo "   • Yellow 'Filter: ...|' when typing in filter mode"
echo "   • Filter controls in footer: 'F: Filter | C: Clear filter'"
echo "   • Filtered count display: 'Showing X/Y endpoints'"
echo ""
echo "✅ User Experience:"
echo "   • 200ms debouncing - smooth typing experience"
echo "   • Alphabetical sorting maintained in filtered results"
echo "   • Easy clear with 'C' key"
echo "   • Enter/Esc to exit filter mode"
echo ""
echo "✅ Integration:"
echo "   • Works alongside Request Log filtering"
echo "   • Help tab documents the new feature"
echo "   • Footer controls show appropriate options per tab"
echo ""
echo "🚀 Start the TUI to test configuration filtering:"
echo "   ./bin/webserver --client"
echo ""
echo "💡 Suggested Test Flow:"
echo "   1. Go to Configuration tab - see all 9 endpoints"
echo "   2. Filter by 'api' - see API endpoints only"
echo "   3. Clear filter with 'C' - see all endpoints again"
echo "   4. Filter by 'delay' - see delay-type endpoints"
echo "   5. Try 'user', 'error', 'admin' filters"
echo "   6. Check Help tab for updated documentation"
echo ""
echo "⚠️  Press ENTER when you're done testing..."
read

# Stop the server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 Server stopped."
echo ""
echo "✨ Configuration Tab Filtering Test Complete!"
echo ""
echo "📋 What Was Implemented:"
echo "  ✅ Configuration filtering state fields in TUI model"
echo "  ✅ Keyboard handling for Configuration tab (F, C keys)"
echo "  ✅ filterConfigEndpoints() function with text matching"
echo "  ✅ Updated Configuration view to show filtered results"
echo "  ✅ Visual filter indicators and status display"
echo "  ✅ Footer controls specific to Configuration tab"
echo "  ✅ Filter debouncing with 200ms delay"
echo "  ✅ Updated help documentation"
echo "  ✅ Alphabetical sorting preserved in filtered results"
echo ""
echo "🎯 Filter Capabilities:"
echo "  🔍 Searches: endpoint paths, types, and error messages"
echo "  ⚡ Real-time: 200ms debouncing for smooth typing"
echo "  🎨 Visual: Green active indicators, yellow input mode"
echo "  📊 Status: Shows 'Showing X/Y endpoints' count"
echo "  🔄 Management: Easy clear with 'C', change with 'F'"
echo ""
echo "🎉 Perfect for navigating large endpoint configurations!" 