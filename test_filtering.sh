#!/bin/bash

echo "🔍 Testing WebServer TUI Request Log Filtering Features..."

# Start the server in the background
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

# Generate diverse requests for filtering demonstrations
echo "Generating diverse request patterns for filtering tests..."

# Create patterns with different endpoints, methods, and IPs
for i in {1..15}; do
    echo "Generating request batch $i/15..."
    
    # Various API endpoints
    curl -s http://localhost:8080/api/error > /dev/null &
    curl -s http://localhost:8080/api/delay > /dev/null &
    curl -s http://localhost:8080/api/flaky > /dev/null &
    
    # Static files
    curl -s http://localhost:8080/ > /dev/null &
    curl -s http://localhost:8080/index.html > /dev/null &
    
    # Configuration and stats (these will be filterable)
    curl -s http://localhost:8080/config > /dev/null &
    curl -s http://localhost:8080/stats > /dev/null &
    
    # Some POST requests for method filtering
    curl -s -X POST -H "Content-Type: application/json" \
         -d '{"test": true}' http://localhost:8080/api/test > /dev/null &
    
    # Some requests that will generate 404s
    curl -s http://localhost:8080/nonexistent > /dev/null &
    curl -s http://localhost:8080/api/missing > /dev/null &
    
    # Brief pause between batches
    sleep 0.3
done

# Wait for requests to complete
echo "Waiting for requests to complete..."
sleep 8

echo ""
echo "✅ Server is running with diverse request log data!"
echo ""
echo "🔍 Advanced Request Log Filtering Features to Test:"
echo ""
echo "   ┌─ Time Ordering & Auto-Update ──────────────────────────────────┐"
echo "   │ • Requests are sorted by time (newest first)                    │"
echo "   │ • Auto-updates every 1 second (faster than before)             │"
echo "   │ • Check timestamps to see chronological ordering               │"
echo "   └──────────────────────────────────────────────────────────────────┘"
echo ""
echo "   ┌─ Stats Endpoint Toggle ─────────────────────────────────────────┐"
echo "   │ • Press 'S' to hide/show /stats and /config requests           │"
echo "   │ • Useful to focus on application traffic vs monitoring         │"
echo "   │ • Toggle state shown in filter indicator                       │"
echo "   └──────────────────────────────────────────────────────────────────┘"
echo ""
echo "   ┌─ Text Filtering with Debouncing ────────────────────────────────┐"
echo "   │ • Press 'F' to enter filter mode                               │"
echo "   │ • Type to search: 'api', 'error', 'POST', '192.168', etc.      │"
echo "   │ • 200ms debouncing - filter applies after you stop typing      │"
echo "   │ • Matching text highlighted in yellow                          │"
echo "   │ • Searches path, method, and IP address                        │"
echo "   │ • Press Enter/Esc to exit filter mode                          │"
echo "   └──────────────────────────────────────────────────────────────────┘"
echo ""
echo "   ┌─ Filter Combinations ───────────────────────────────────────────┐"
echo "   │ • Use both text filter AND stats toggle together               │"
echo "   │ • Example: Filter 'api' + Hide stats = only API endpoints      │"
echo "   │ • Filter count shows: 'Showing X/Y requests'                   │"
echo "   └──────────────────────────────────────────────────────────────────┘"
echo ""
echo "🎯 Start the TUI to test filtering:"
echo "   ./bin/webserver --client"
echo ""
echo "📋 Filtering Test Scenarios:"
echo "   1. Go to Request Log tab"
echo "   2. Press 'S' to toggle internal endpoints visibility"
echo "   3. Press 'F' and type 'api' - see only API endpoints"
echo "   4. Clear with 'C' and try 'POST' - see only POST requests"
echo "   5. Try 'error' - see only error endpoint requests"
echo "   6. Try '192.168' - see requests from generated IP ranges"
echo "   7. Notice yellow highlighting on matching text"
echo "   8. Check filter indicators below tabs"
echo "   9. Observe real-time updates every 1 second"
echo "   10. Try combining filters (S + F together)"
echo ""
echo "💡 What to Look For:"
echo "   • Yellow typing cursor in filter mode"
echo "   • Green filter status indicators"
echo "   • Yellow highlighted matching text"
echo "   • Filtered counts: 'Showing 5/20 requests'"
echo "   • Real-time updates without losing filter state"
echo "   • Sorted timestamps (newest at top)"
echo ""
echo "🔍 API Tests (generate more traffic while TUI is open):"
echo "   curl http://localhost:8080/api/error"
echo "   curl http://localhost:8080/stats"
echo "   curl -X POST http://localhost:8080/api/test"
echo ""
echo "⚠️  Press ENTER when you're done testing the filtering features..."
read

# Stop the server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 Server stopped."
echo ""
echo "✨ Request Log Filtering Test Complete!"
echo ""
echo "🎉 Features Implemented:"
echo "   ✅ Time-ordered requests (newest first)"
echo "   ✅ 1-second auto-refresh (faster updates)"
echo "   ✅ Stats endpoint toggle (S key)"
echo "   ✅ Text filtering with debouncing (F key)" 
echo "   ✅ 200ms debounce timing"
echo "   ✅ Text highlighting of matches"
echo "   ✅ Multi-field search (path, method, IP)"
echo "   ✅ Filter combination support"
echo "   ✅ Visual filter indicators"
echo "   ✅ Real-time filter count display"
echo "   ✅ Enhanced help documentation"
echo ""
echo "🔧 Filter Controls Summary:"
echo "   • F - Enter/exit filter mode"
echo "   • S - Toggle hide /stats requests"
echo "   • C - Clear all filters"
echo "   • Enter/Esc - Exit filter mode"
echo "   • All changes apply with 200ms debouncing" 