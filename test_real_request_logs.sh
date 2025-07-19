#!/bin/bash

echo "📋 Testing Real Request Logging..."
echo ""
echo "✅ Major Improvements:"
echo "  • Real request data instead of fake/artificial data"
echo "  • Original request timestamps (when request was made)"
echo "  • Actual request durations (how long it took to process)"
echo "  • Ordered by actual request time (newest first)"
echo "  • All requests logged with middleware"
echo ""

# Start the server
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Making real requests to generate authentic log data..."

# Make various requests with different timing patterns
echo "Making quick requests..."
curl -s http://localhost:8080/ > /dev/null
curl -s http://localhost:8080/stats > /dev/null
curl -s http://localhost:8080/config > /dev/null

sleep 1

echo "Making error requests..."
curl -s http://localhost:8080/api/error > /dev/null
curl -s http://localhost:8080/nonexistent > /dev/null

sleep 1

echo "Making delay requests (these will show longer durations)..."
curl -s http://localhost:8080/api/delay > /dev/null &
curl -s http://localhost:8080/api/delay > /dev/null &

# Make more requests while delay is running
curl -s http://localhost:8080/stats > /dev/null
curl -s http://localhost:8080/api/flaky > /dev/null

# Wait for delay requests to complete
sleep 3

echo "Making final batch of requests..."
curl -s -X POST http://localhost:8080/api/test > /dev/null
curl -s http://localhost:8080/config > /dev/null

sleep 1

echo ""
echo "🎯 Test Real Request Logging:"
echo ""
echo "┌─ Real Request Data ───────────────────────────────────────────────┐"
echo "│ • Go to Request Log tab in TUI                                    │"
echo "│ • Check timestamps - they should be REAL request times            │"
echo "│ • Check durations:                                                 │"
echo "│   - Regular requests: 1-50ms                                      │"
echo "│   - Delay requests: ~2000ms (from /api/delay endpoint)            │"
echo "│   - Error requests: various durations                             │"
echo "│ • Order: Newest requests at top (by actual request time)          │"
echo "│ • All request data is captured from actual server requests        │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ What to Verify ─────────────────────────────────────────────────┐"
echo "│ 1. Request timestamps match when you made the requests            │"
echo "│ 2. Delay endpoints show ~2000ms duration                          │"
echo "│ 3. Regular endpoints show realistic durations (1-50ms)            │"
echo "│ 4. Ordering is chronological (newest at top)                     │"
echo "│ 5. All HTTP methods, paths, status codes are real                 │"
echo "│ 6. IP addresses are real client addresses                         │"
echo "│ 7. No more artificial/fake timestamp changes                      │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "🧪 Test the API Endpoint:"
echo "   curl http://localhost:8080/requestlog | jq"
echo ""
echo "💡 Make more requests while TUI is open to see real-time updates:"
echo "   curl http://localhost:8080/api/error"
echo "   curl http://localhost:8080/api/delay"
echo "   curl http://localhost:8080/stats"
echo ""
echo "🚀 Start the TUI to see real request logs:"
echo "   ./bin/webserver --client"
echo ""
echo "⚠️  Press ENTER when you're done testing..."
read

# Stop the server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 Server stopped."
echo ""
echo "✨ Real Request Logging Test Complete!"
echo ""
echo "📋 What Was Implemented:"
echo "  ✅ Server-side request logging middleware"
echo "  ✅ Real timestamp capture (when request started)"
echo "  ✅ Actual duration measurement (how long it took)"
echo "  ✅ Thread-safe request log storage (last 1000 requests)"
echo "  ✅ New /requestlog API endpoint"
echo "  ✅ TUI fetches real data instead of generating fake data"
echo "  ✅ Proper ordering by actual request time"
echo "  ✅ Real status codes, methods, paths, and IP addresses"
echo ""
echo "🎉 Now you have authentic request logging with real timing data!" 