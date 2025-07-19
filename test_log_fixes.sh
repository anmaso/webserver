#!/bin/bash

echo "🔧 Testing Request Log Fixes..."
echo ""
echo "✅ Fixes Implemented:"
echo "  1. Fixed text disappearing when filtering (truncate before highlighting)"
echo "  2. Improved time ordering with realistic timestamps"
echo "  3. Enhanced highlight function robustness"
echo ""

# Start the server in the background
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Generate some requests to create log entries
echo "Generating sample requests..."
for i in {1..8}; do
    curl -s http://localhost:8080/api/error > /dev/null &
    curl -s http://localhost:8080/api/delay > /dev/null &
    curl -s http://localhost:8080/stats > /dev/null &
    curl -s http://localhost:8080/ > /dev/null &
    sleep 0.5
done

# Wait for requests to complete
sleep 4

echo ""
echo "🧪 Test the Following Fixes:"
echo ""
echo "┌─ Time Ordering Fix ──────────────────────────────────────────────┐"
echo "│ • Requests should be sorted by timestamp (newest first)          │"
echo "│ • Check the Time column - should go from most recent to oldest   │"
echo "│ • Look for the ⏰ icon indicating time sorting                   │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Text Highlighting Fix ──────────────────────────────────────────┐"
echo "│ • Go to Request Log tab, press 'F', type 'api'                   │"
echo "│ • Endpoint paths should NOT disappear when filtering             │"
echo "│ • Text should be highlighted in yellow without cutting off       │"
echo "│ • Try filtering 'error', 'stats', 'delay' - text should remain   │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Additional Tests ───────────────────────────────────────────────┐"
echo "│ • Try filtering by method: 'GET', 'POST'                         │"
echo "│ • Try filtering by IP: '127.0.0.1', '192.168'                    │"
echo "│ • Use 'S' to toggle internal endpoints visibility                │"
echo "│ • Combine filters (text + stats toggle)                          │"
echo "│ • Check that timestamps are properly sorted after filtering      │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "🚀 Start the TUI to test the fixes:"
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
echo "✨ Request Log Fixes Tested!"
echo ""
echo "📋 Summary of What Was Fixed:"
echo "  ✅ Text no longer disappears when filtering"
echo "  ✅ Proper time ordering (newest → oldest)"
echo "  ✅ Robust highlighting that preserves text"
echo "  ✅ Better timestamp generation for realistic ordering"
echo "  ✅ Improved filtering UX with visual indicators" 