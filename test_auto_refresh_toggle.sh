#!/bin/bash

echo "🔄 Testing Auto-Refresh Toggle Feature..."
echo ""
echo "✅ New Feature Added:"
echo "  • Auto-refresh toggle (A key) to enable/disable automatic updates"
echo "  • Scrolling automatically disables auto-refresh to prevent disruption"
echo "  • Status displayed in header: 'auto-refresh: ON/OFF'"
echo "  • Manual refresh (R key) works when auto-refresh is disabled"
echo ""

# Start the server
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Making initial requests to generate log data..."
curl -s http://localhost:8080/ > /dev/null
curl -s http://localhost:8080/stats > /dev/null
curl -s http://localhost:8080/config > /dev/null

sleep 2

echo ""
echo "🎯 Test Auto-Refresh Toggle:"
echo ""
echo "┌─ Basic Auto-Refresh Testing ─────────────────────────────────────┐"
echo "│ 1. Start TUI: ./bin/webserver --client                              │"
echo "│ 2. Go to Request Log tab                                          │"
echo "│ 3. Notice header shows: 'auto-refresh: ON'                       │"
echo "│ 4. Watch requests update automatically every 1 second            │"
echo "│ 5. Press 'A' to toggle auto-refresh OFF                          │"
echo "│ 6. Header should change to: 'auto-refresh: OFF'                  │"
echo "│ 7. Observe that new requests DON'T appear automatically          │"
echo "│ 8. Press 'R' to manually refresh and see new data                │"
echo "│ 9. Press 'A' again to turn auto-refresh back ON                  │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Scroll-Disable Testing ─────────────────────────────────────────┐"
echo "│ 1. Ensure auto-refresh is ON (header shows 'auto-refresh: ON')   │"
echo "│ 2. Use any scroll key: ↑↓, j/k, Page Up/Down, Home/End           │"
echo "│ 3. Header should immediately change to: 'auto-refresh: OFF'      │"
echo "│ 4. Auto-updates should stop - scroll position stays stable       │"
echo "│ 5. Press 'A' to manually turn auto-refresh back ON               │"
echo "│ 6. New requests should start appearing again                      │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "🧪 Generate Test Traffic While Testing:"

# Function to generate continuous traffic
generate_traffic() {
    while true; do
        curl -s http://localhost:8080/ > /dev/null &
        curl -s http://localhost:8080/api/error > /dev/null &
        curl -s http://localhost:8080/stats > /dev/null &
        curl -s http://localhost:8080/config > /dev/null &
        sleep 2
        
        curl -s http://localhost:8080/api/delay > /dev/null &
        curl -s http://localhost:8080/api/flaky > /dev/null &
        sleep 3
    done
}

# Start background traffic generation
echo "Starting continuous test traffic..."
generate_traffic &
TRAFFIC_PID=$!

echo ""
echo "🔧 Features to Test:"
echo ""
echo "✅ Auto-Refresh Controls:"
echo "   • A key toggles auto-refresh on/off"
echo "   • Header shows current status (ON/OFF)"
echo "   • Manual refresh (R key) works when OFF"
echo ""
echo "✅ Scroll Auto-Disable:"
echo "   • Any scroll action disables auto-refresh"
echo "   • Prevents view jumping during manual examination"
echo "   • User maintains control over scroll position"
echo ""
echo "✅ UI Indicators:"
echo "   • Footer shows 'A: Auto-refresh' control"
echo "   • Header displays current auto-refresh status"
echo "   • Help tab documents the new feature"
echo ""
echo "🚀 Start the TUI to test auto-refresh toggle:"
echo "   ./bin/webserver --client"
echo ""
echo "💡 Test Scenarios:"
echo "   1. Toggle auto-refresh and observe behavior changes"
echo "   2. Scroll and watch auto-refresh disable automatically"
echo "   3. Use manual refresh when auto-refresh is disabled"
echo "   4. Check help tab for documentation"
echo ""
echo "⚠️  Press ENTER when you're done testing..."
read

# Stop background traffic and server
echo "Stopping test traffic..."
kill $TRAFFIC_PID 2>/dev/null
wait $TRAFFIC_PID 2>/dev/null

echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 All processes stopped."
echo ""
echo "✨ Auto-Refresh Toggle Test Complete!"
echo ""
echo "📋 What Was Implemented:"
echo "  ✅ Auto-refresh toggle (A key) for Request Log tab"
echo "  ✅ Automatic disable when user scrolls manually"
echo "  ✅ Visual status indicator in header (ON/OFF)"
echo "  ✅ Footer controls updated to show 'A: Auto-refresh'"
echo "  ✅ Help documentation with detailed explanations"
echo "  ✅ Smart UX: preserves user intent when examining logs"
echo "  ✅ Manual refresh still works when auto-refresh is off"
echo ""
echo "🎯 Benefits:"
echo "  🔄 Better control over when data updates"
echo "  📍 Stable view when examining specific entries"
echo "  ⚡ Reduced network requests when disabled"
echo "  🎨 Clear visual feedback of current state"
echo ""
echo "🎉 Perfect for examining historical requests without interruption!" 