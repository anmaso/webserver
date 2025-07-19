#!/bin/bash

echo "🔤 Testing Alphabetical Ordering in TUI..."
echo ""
echo "✅ Stability Improvements:"
echo "  • All endpoint listings now display in alphabetical order"
echo "  • No more flickering caused by random map iteration order"
echo "  • Consistent display across refreshes"
echo "  • Applied to Overview, Configuration, and Statistics tabs"
echo ""

# Start the server
echo "Starting server..."
./bin/webserver &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Creating test configuration with multiple endpoints in mixed order..."

# Add endpoints in mixed alphabetical order to test sorting
curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/zebra", "config": {"type": "error", "status_code": 500, "message": "Zebra endpoint"}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/alpha", "config": {"type": "delay", "delay_ms": 100}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/beta", "config": {"type": "conditional_error", "error_every_n": 3, "status_code": 404}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/gamma", "config": {"type": "error", "status_code": 400, "message": "Gamma error"}}' > /dev/null

curl -s -X POST http://localhost:8080/config \
  -H 'Content-Type: application/json' \
  -d '{"operation": "add", "path": "/delta", "config": {"type": "delay", "delay_ms": 200}}' > /dev/null

sleep 2

echo "Making requests to generate statistics in mixed order..."
curl -s http://localhost:8080/zebra > /dev/null
curl -s http://localhost:8080/alpha > /dev/null
curl -s http://localhost:8080/beta > /dev/null
curl -s http://localhost:8080/gamma > /dev/null
curl -s http://localhost:8080/delta > /dev/null
curl -s http://localhost:8080/beta > /dev/null
curl -s http://localhost:8080/alpha > /dev/null
curl -s http://localhost:8080/zebra > /dev/null

sleep 1

echo ""
echo "🎯 Test Alphabetical Ordering:"
echo ""
echo "┌─ Overview Tab ────────────────────────────────────────────────────┐"
echo "│ 1. Go to Overview tab                                              │"
echo "│ 2. Check 'Endpoint Summary' section                               │"
echo "│ 3. Endpoints should appear in order:                              │"
echo "│    • /alpha (delay)                                               │"
echo "│    • /beta (conditional_error)                                    │"
echo "│    • /delta (delay)                                               │"
echo "│    • /gamma (error)                                               │"
echo "│    • /zebra (error)                                               │"
echo "│ 4. Press 'R' multiple times - order should stay the same!         │"
echo "└────────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Configuration Tab ───────────────────────────────────────────────┐"
echo "│ 1. Go to Configuration tab                                        │"
echo "│ 2. Check 'Configured Endpoints' section                           │"
echo "│ 3. Endpoints should appear in alphabetical order:                 │"
echo "│    • /alpha (with delay configuration)                            │"
echo "│    • /beta (with conditional error configuration)                 │"
echo "│    • /delta (with delay configuration)                            │"
echo "│    • /gamma (with error configuration)                            │"
echo "│    • /zebra (with error configuration)                            │"
echo "│ 4. Press 'R' multiple times - order should never change!          │"
echo "└────────────────────────────────────────────────────────────────────┘"
echo ""
echo "┌─ Statistics Tab ──────────────────────────────────────────────────┐"
echo "│ 1. Go to Statistics tab                                           │"
echo "│ 2. Check 'Per-Endpoint Statistics' section                        │"
echo "│ 3. Endpoint stats should appear in alphabetical order:            │"
echo "│    ━━━ /alpha ━━━                                                 │"
echo "│    ━━━ /beta ━━━                                                  │"
echo "│    ━━━ /delta ━━━                                                 │"
echo "│    ━━━ /gamma ━━━                                                 │"
echo "│    ━━━ /zebra ━━━                                                 │"
echo "│ 4. Press 'R' multiple times - no flickering!                      │"
echo "└────────────────────────────────────────────────────────────────────┘"
echo ""
echo "🧪 Additional Testing:"
echo ""

# Function to generate more requests
generate_mixed_requests() {
    for i in {1..5}; do
        curl -s http://localhost:8080/zebra > /dev/null &
        curl -s http://localhost:8080/gamma > /dev/null &
        curl -s http://localhost:8080/beta > /dev/null &
        curl -s http://localhost:8080/delta > /dev/null &
        curl -s http://localhost:8080/alpha > /dev/null &
        sleep 1
    done
}

echo "Generating additional mixed traffic..."
generate_mixed_requests &
TRAFFIC_PID=$!

echo ""
echo "🔧 What to Verify:"
echo ""
echo "✅ Stable Display:"
echo "   • Endpoint order never changes between refreshes"
echo "   • No visual flickering or jumping text"
echo "   • Alphabetical sorting: /alpha, /beta, /delta, /gamma, /zebra"
echo ""
echo "✅ Cross-Tab Consistency:"
echo "   • Same alphabetical order across all three tabs"
echo "   • Overview summary matches Configuration and Statistics order"
echo ""
echo "✅ Refresh Stability:"
echo "   • Press 'R' repeatedly - order remains consistent"
echo "   • Auto-refresh (if enabled) doesn't cause flickering"
echo "   • Statistics update values but maintain endpoint order"
echo ""
echo "🚀 Start the TUI to test alphabetical ordering:"
echo "   ./bin/webserver --client"
echo ""
echo "⚠️  Press ENTER when you're done testing..."
read

# Stop traffic generation and server
echo "Stopping test traffic..."
kill $TRAFFIC_PID 2>/dev/null
wait $TRAFFIC_PID 2>/dev/null

echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🛑 All processes stopped."
echo ""
echo "✨ Alphabetical Ordering Test Complete!"
echo ""
echo "📋 What Was Fixed:"
echo "  ✅ Overview Tab - Endpoint Summary sorted alphabetically"
echo "  ✅ Configuration Tab - Configured Endpoints sorted alphabetically"  
echo "  ✅ Statistics Tab - Per-Endpoint Statistics sorted alphabetically"
echo "  ✅ All endpoint listings now use deterministic ordering"
echo "  ✅ No more flickering caused by Go's random map iteration"
echo "  ✅ Consistent display across all refreshes"
echo ""
echo "🎯 Technical Implementation:"
echo "  📝 Collect endpoint paths from maps into slices"
echo "  🔤 Sort paths alphabetically using sort.Strings()"
echo "  🔄 Iterate in sorted order instead of map order"
echo "  ⚡ Applied to all three main TUI tabs"
echo ""
echo "🎉 Now the TUI provides a stable, professional display experience!" 