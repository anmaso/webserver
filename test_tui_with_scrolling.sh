#!/bin/bash

echo "🚀 Testing WebServer TUI with Scrolling and Request Logs..."

# Start the server in the background
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

# Make many requests to generate lots of data for scrolling
echo "Making multiple requests to generate scrollable content..."

# Generate a variety of requests
for i in {1..10}; do
    echo "Making request batch $i..."
    
    # Various endpoint requests
    curl -s http://localhost:8080/api/error > /dev/null &
    curl -s http://localhost:8080/api/delay > /dev/null &
    curl -s http://localhost:8080/api/flaky > /dev/null
    curl -s http://localhost:8080/api/flaky > /dev/null
    curl -s http://localhost:8080/api/flaky > /dev/null
    curl -s http://localhost:8080/ > /dev/null
    curl -s http://localhost:8080/config > /dev/null
    curl -s http://localhost:8080/stats > /dev/null
    
    # Small delay between batches
    sleep 0.5
done

# Wait for all background requests to complete
echo "Waiting for requests to complete..."
sleep 5

echo ""
echo "✅ Server is running with lots of test data!"
echo ""
echo "📊 TUI Features to Test:"
echo "   ┌─ Scrolling Functionality ─────────────────────────────────┐"
echo "   │ • ↑↓ or j/k keys    - Scroll line by line                 │"
echo "   │ • Page Up/Down or u/d - Scroll by half page               │"
echo "   │ • Home/End or g/G   - Jump to top/bottom                  │"
echo "   │ • Tab/Shift+Tab     - Switch tabs (each has own scroll)   │"
echo "   │ • Look for ▲▼ scroll indicators when content is long      │"
echo "   │ • Footer shows scroll position: 'Scroll: X/Y'             │"
echo "   └────────────────────────────────────────────────────────────┘"
echo ""
echo "   ┌─ Tab Content to Explore ──────────────────────────────────┐"
echo "   │ • Overview      - Server info + recent activity           │"
echo "   │ • Configuration - Detailed endpoint configs + API info    │"
echo "   │ • Statistics    - Comprehensive per-endpoint metrics      │"
echo "   │ • Request Log   - Generated request entries with colors   │"
echo "   │ • Help          - Very long help content (tests scrolling)│"
echo "   └────────────────────────────────────────────────────────────┘"
echo ""
echo "🎯 Start the TUI now to see scrolling in action:"
echo "   ./bin/webserver --client"
echo ""
echo "💡 Things to try in the TUI:"
echo "   1. Switch to Help tab - it has lots of content to scroll through"
echo "   2. Use ↑↓ keys to scroll line by line"
echo "   3. Use Page Up/Down for faster scrolling"
echo "   4. Check Statistics tab - should have detailed endpoint data"
echo "   5. Look at Request Log tab - should show many colored entries"
echo "   6. Notice scroll indicators (▲▼) when there's more content"
echo "   7. Try resizing terminal to see responsive scrolling"
echo ""
echo "🔍 Quick API tests:"
echo "   curl http://localhost:8080/config | jq"
echo "   curl http://localhost:8080/stats | jq"
echo ""
echo "⚠️  Press ENTER to stop the server when done testing..."
read

# Stop the server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 Server stopped."
echo "✨ TUI scrolling test complete!"
echo ""
echo "📝 What we implemented:"
echo "   • Full scrolling support in all tabs"
echo "   • Per-tab scroll position memory"
echo "   • Vim-style navigation (j/k/g/G/u/d)"
echo "   • Scroll indicators and position display"
echo "   • Responsive content that adapts to terminal size"
echo "   • Request log with sample data generation"
echo "   • Enhanced help system with detailed scrolling instructions" 