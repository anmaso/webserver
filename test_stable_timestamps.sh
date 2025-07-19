#!/bin/bash

echo "🕐 Testing Stable Timestamps Fix..."
echo ""
echo "❌ Previous Problem:"
echo "   • Time column values were changing every second"
echo "   • Timestamps kept moving forward because they were"
echo "     regenerated using time.Now() every refresh"
echo ""
echo "✅ Fix Applied:"
echo "   • Request log data is now cached after first generation"
echo "   • Timestamps use a fixed base time, so they stay stable"
echo "   • Press 'R' in Request Log tab to generate fresh timestamps"
echo ""

# Start the server
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Generate some requests
echo "Generating sample requests..."
for i in {1..5}; do
    curl -s http://localhost:8080/api/error > /dev/null &
    curl -s http://localhost:8080/stats > /dev/null &
    sleep 0.3
done

sleep 2

echo ""
echo "🧪 Test Instructions:"
echo ""
echo "1. Start the TUI: ./bin/webserver --client"
echo "2. Go to Request Log tab"
echo "3. Watch the Time column for 10+ seconds"
echo "4. ✅ Time values should NOT change anymore"
echo "5. Press 'R' to refresh and get new timestamps"
echo "6. Time values will update once, then stay stable again"
echo ""
echo "💡 What you should observe:"
echo "   • Timestamps remain constant during auto-refresh"
echo "   • Only change when you manually press 'R'"
echo "   • Header shows 'stable timestamps' indicator"
echo ""
echo "⚠️  Press ENTER to stop the server when done testing..."
read

# Stop the server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 Server stopped."
echo ""
echo "✨ Stable Timestamps Fix Complete!"
echo ""
echo "📝 Summary:"
echo "  ✅ Timestamps now remain stable between refreshes"
echo "  ✅ Added caching to prevent regeneration every second" 
echo "  ✅ 'R' key in Request Log tab generates fresh timestamps"
echo "  ✅ Better user experience with predictable time values" 