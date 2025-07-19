#!/bin/bash

echo "🐛 Testing Request Logging Fixes..."
echo ""
echo "✅ Issues Being Fixed:"
echo "  • Show complete request query/URI (not just endpoint path)"
echo "  • Remove duplicate log entries in console and TUI"
echo "  • Proper request URI capture with query parameters"
echo ""

# Start the server
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Making requests with query parameters to test complete URI logging..."

# Make various requests with query parameters
curl -s "http://localhost:8080/?page=1&limit=10" > /dev/null
curl -s "http://localhost:8080/stats?format=json&detailed=true" > /dev/null
curl -s "http://localhost:8080/config?version=latest" > /dev/null
curl -s "http://localhost:8080/api/error?code=500&message=test" > /dev/null

# Add some endpoints to test
curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/search", "config": {"type": "delay", "delay_ms": 100}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/api/test", "config": {"type": "error", "status_code": 404}}' > /dev/null

sleep 1

# Make requests to the new endpoints with query parameters
curl -s "http://localhost:8080/api/search?q=golang&page=1&sort=date" > /dev/null
curl -s "http://localhost:8080/api/test?debug=true&trace=enabled" > /dev/null
curl -s "http://localhost:8080/api/search?q=web+server&limit=50" > /dev/null

sleep 1

echo ""
echo "🎯 Test Request Logging Fixes:"
echo ""
echo "┌─ Complete Request URI Testing ───────────────────────────────────┐"
echo "│ 1. Start TUI: ./bin/webserver --client                              │"
echo "│ 2. Go to Request Log tab                                          │"
echo "│ 3. Check recent requests show COMPLETE URIs:                     │"
echo "│    • /?page=1&limit=10 (not just /)                             │"
echo "│    • /stats?format=json&detailed=true (not just /stats)         │"
echo "│    • /config?version=latest (not just /config)                  │"
echo "│    • /api/search?q=golang&page=1&sort=date (full query)         │"
echo "│    • /api/test?debug=true&trace=enabled (full query)            │"
echo "│ 4. All query parameters should be visible in the Path column     │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Duplicate Logging Testing ──────────────────────────────────────┐"
echo "│ 1. Check server console output (this terminal)                   │"
echo "│ 2. Each request should appear ONCE in console                    │"
echo "│ 3. In TUI Request Log, each request should appear ONCE           │"
echo "│ 4. No duplicate entries with same timestamp/method/path          │"
echo "│ 5. Count requests: should match actual requests made             │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""

# Generate some more diverse requests to test
echo "Making additional requests to verify fixes..."

curl -s "http://localhost:8080/?test=query&another=param" > /dev/null
curl -s "http://localhost:8080/api/search?q=test+query&filter=recent&max=100" > /dev/null
curl -s "http://localhost:8080/stats?format=detailed&include=all" > /dev/null

# Make requests to non-existent endpoints with query params
curl -s "http://localhost:8080/nonexistent?param=value&debug=1" > /dev/null

sleep 2

echo ""
echo "🔧 What to Verify:"
echo ""
echo "✅ Complete Request URI Display:"
echo "   • Query parameters visible in TUI Request Log"
echo "   • Console log shows full URI: GET /?page=1&limit=10"
echo "   • Path column shows complete request including ?param=value"
echo ""
echo "✅ No Duplicate Entries:"
echo "   • Each request appears exactly once in console"
echo "   • Each request appears exactly once in TUI"
echo "   • No duplicate timestamps for same request"
echo "   • Total count matches actual requests made"
echo ""
echo "✅ Console Output Example:"
echo "   GET /?page=1&limit=10 ::1:port"
echo "   GET /stats?format=json&detailed=true ::1:port"
echo "   GET /api/search?q=golang&page=1&sort=date ::1:port"
echo ""
echo "✅ TUI Request Log Example:"
echo "   Time     Method  Path                              Status  Duration"
echo "   15:04:23 GET     /?page=1&limit=10                200     15ms"
echo "   15:04:24 GET     /stats?format=json&detailed=true  200     8ms"
echo "   15:04:25 GET     /api/search?q=golang&page=1...    100     95ms"
echo ""
echo "🚀 Start the TUI to test request logging fixes:"
echo "   ./bin/webserver --client"
echo ""
echo "💡 Additional Test Commands (run while TUI is open):"
echo "   curl 'http://localhost:8080/stats?detailed=1&format=json'"
echo "   curl 'http://localhost:8080/api/search?q=test&sort=date'"
echo "   curl 'http://localhost:8080/?page=2&limit=20&sort=name'"
echo ""
echo "⚠️  Press ENTER when you're done testing..."
read

# Stop the server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 Server stopped."
echo ""
echo "✨ Request Logging Fixes Test Complete!"
echo ""
echo "📋 What Was Fixed:"
echo "  ✅ Changed r.URL.Path to r.URL.RequestURI() in request logging"
echo "  ✅ Removed duplicate logging from handleRequest function"
echo "  ✅ Updated console logging to show complete request URI"
echo "  ✅ Maintained single logging point through middleware"
echo "  ✅ Fixed both TUI and console output to show query parameters"
echo ""
echo "🎯 Technical Changes:"
echo "  🔧 server.go middleware: Path: r.URL.RequestURI()"
echo "  🔧 handlers.go logRequest: log.Printf with r.URL.RequestURI()"
echo "  🔧 handlers.go broadcastRequestLog: Path: r.URL.RequestURI()"
echo "  🚫 Removed duplicate s.broadcastRequestLog() call from handleRequest"
echo "  📈 Statistics still use r.URL.Path (for proper endpoint grouping)"
echo ""
echo "🎉 Now request logs show complete URIs without duplicates!" 