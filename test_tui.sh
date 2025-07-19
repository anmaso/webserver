#!/bin/bash

echo "🚀 Testing WebServer TUI with real server data..."

# Start the server in the background
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Make some requests to generate data
echo "Making test requests to generate data..."
curl -s http://localhost:8080/api/error > /dev/null
curl -s http://localhost:8080/api/delay > /dev/null &
curl -s http://localhost:8080/api/flaky > /dev/null
curl -s http://localhost:8080/api/flaky > /dev/null
curl -s http://localhost:8080/api/flaky > /dev/null
curl -s http://localhost:8080/ > /dev/null

# Wait for requests to complete
sleep 3

echo "✅ Server is running with test data!"
echo "📊 You can now test the TUI with real data:"
echo "   ./bin/webserver --client"
echo ""
echo "💡 The TUI now includes:"
echo "   - Overview tab with server info and recent activity"
echo "   - Configuration tab with current settings"
echo "   - Statistics tab with real metrics"
echo "   - Request Log tab (simulated for now)"
echo "   - Help tab with keyboard shortcuts and info"
echo ""
echo "⌨️  TUI Controls:"
echo "   - Tab/Shift+Tab: Switch tabs"
echo "   - R: Refresh data"
echo "   - Q: Quit"
echo ""
echo "🔍 Test the configuration API:"
echo "   curl http://localhost:8080/config"
echo "   curl http://localhost:8080/stats"
echo ""
echo "⚠️  Press any key to stop the server..."
read -n 1

# Stop the server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 Server stopped."
echo "✨ TUI test setup complete!" 