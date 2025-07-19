#!/bin/bash

echo "📊 Testing Sorted Log Summaries..."
echo ""
echo "✅ Improvements Applied:"
echo "  • Status Code Distribution - sorted numerically (200, 404, 500)"
echo "  • HTTP Methods - sorted alphabetically (DELETE, GET, POST, PUT)"
echo "  • Applied to both Request Log and Statistics tabs"
echo ""

# Start the server
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Generate diverse requests with different methods and status codes
echo "Generating diverse requests to test sorting..."
for i in {1..3}; do
    # Various methods and endpoints to create diverse status codes
    curl -s http://localhost:8080/api/error > /dev/null &          # 500 status
    curl -s http://localhost:8080/ > /dev/null &                  # 200 status
    curl -s http://localhost:8080/nonexistent > /dev/null &       # 404 status
    curl -s http://localhost:8080/stats > /dev/null &             # 200 status
    curl -s -X POST http://localhost:8080/api/test > /dev/null &  # POST method
    curl -s -X PUT http://localhost:8080/config > /dev/null &     # PUT method
    curl -s -X DELETE http://localhost:8080/api/fake > /dev/null & # DELETE method
    sleep 0.5
done

# Wait for requests to complete
sleep 4

echo ""
echo "🎯 Test the Sorted Summaries:"
echo ""
echo "┌─ Request Log Tab ─────────────────────────────────────────────────┐"
echo "│ 1. Go to Request Log tab                                          │"
echo "│ 2. Scroll to bottom to see 📊 Log Summary                        │"
echo "│ 3. Status Code Distribution should show:                          │"
echo "│    • 200: X entries                                               │"
echo "│    • 404: X entries                                               │"
echo "│    • 500: X entries                                               │"
echo "│    (Sorted numerically, ascending)                                │"
echo "│ 4. HTTP Methods should show (if multiple methods):                │"
echo "│    • DELETE: X entries                                            │"
echo "│    • GET: X entries                                               │"
echo "│    • POST: X entries                                              │"
echo "│    • PUT: X entries                                               │"
echo "│    (Sorted alphabetically)                                        │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Statistics Tab ──────────────────────────────────────────────────┐"
echo "│ 1. Go to Statistics tab                                           │"
echo "│ 2. Check each endpoint's Status Code Distribution                 │"
echo "│ 3. Status codes should be sorted numerically                      │"
echo "│ 4. Compare with previous random order - now consistent!           │"
echo "└───────────────────────────────────────────────────────────────────┘"
echo ""
echo "🧪 Additional Testing:"
echo "   • Try filtering in Request Log tab to see sorted filtered results"
echo "   • Use different filters to generate various status code mixes"
echo "   • Check that sorting works with any combination of codes/methods"
echo ""
echo "🚀 Start the TUI to test sorted summaries:"
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
echo "✨ Sorted Log Summaries Test Complete!"
echo ""
echo "📋 What Was Improved:"
echo "  ✅ Status codes now display in numerical order (200, 404, 500)"
echo "  ✅ HTTP methods now display alphabetically (DELETE, GET, POST, PUT)"
echo "  ✅ Consistent ordering across Request Log and Statistics tabs"
echo "  ✅ Better readability and predictable data presentation"
echo "  ✅ Professional, organized summary displays" 