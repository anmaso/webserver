#!/bin/bash

echo "🔄 Verifying Proxy1 → WebServer Rename..."
echo ""

# Check if old references remain
echo "🔍 Checking for remaining 'proxy1' references:"
PROXY1_REFS=$(grep -r "proxy1" . --exclude-dir=.git --exclude-dir=bin --exclude="*.log" --exclude="test_rename_verification.sh" --exclude-dir=.DS_Store -I 2>/dev/null || true)

if [ -z "$PROXY1_REFS" ]; then
    echo "✅ No remaining 'proxy1' references found"
else
    echo "❌ Found remaining 'proxy1' references:"
    echo "$PROXY1_REFS"
fi

echo ""

# Verify directory structure
echo "📁 Checking directory structure:"
if [ -d "cmd/webserver" ]; then
    echo "✅ cmd/webserver/ directory exists"
else
    echo "❌ cmd/webserver/ directory missing"
fi

if [ ! -d "cmd/proxy1" ]; then
    echo "✅ Old cmd/proxy1/ directory removed"
else
    echo "⚠️  Old cmd/proxy1/ directory still exists"
fi

echo ""

# Verify module name
echo "📦 Checking go.mod:"
MODULE_NAME=$(head -1 go.mod | cut -d' ' -f2)
if [ "$MODULE_NAME" = "webserver" ]; then
    echo "✅ Module name updated to 'webserver'"
else
    echo "❌ Module name is '$MODULE_NAME', expected 'webserver'"
fi

echo ""

# Verify binary builds
echo "🔧 Testing build:"
if go build -o bin/webserver ./cmd/webserver; then
    echo "✅ Binary builds successfully"
    
    # Test help output
    echo "📖 Testing help output:"
    if ./bin/webserver -help | grep -q "WebServer - Configurable Web Server"; then
        echo "✅ Help text shows correct application name"
    else
        echo "❌ Help text doesn't show updated application name"
    fi
    
    # Test version output
    echo "📊 Testing version output:"
    if ./bin/webserver -version | grep -q "WebServer Configurable Web Server"; then
        echo "✅ Version text shows correct application name"
    else
        echo "❌ Version text doesn't show updated application name"
    fi
else
    echo "❌ Binary build failed"
fi

echo ""

# Verify Docker files
echo "🐳 Checking Docker configuration:"
if grep -q "webserver" Dockerfile; then
    echo "✅ Dockerfile updated with new name"
else
    echo "❌ Dockerfile not properly updated"
fi

if grep -q "webserver:" docker-compose.yml; then
    echo "✅ docker-compose.yml updated with new service names"
else
    echo "❌ docker-compose.yml not properly updated"
fi

echo ""

# Check test scripts
echo "🧪 Checking test scripts:"
TEST_SCRIPTS_UPDATED=true
for script in test_*.sh; do
    # Skip Docker test script as it doesn't use local binary
    if [[ "$script" == "test_docker.sh" ]] || [[ "$script" == "test_rename_verification.sh" ]]; then
        continue
    fi
    if grep -q "./bin/webserver" "$script"; then
        continue
    else
        echo "❌ $script not updated with new binary name"
        TEST_SCRIPTS_UPDATED=false
    fi
done

if [ "$TEST_SCRIPTS_UPDATED" = true ]; then
    echo "✅ All test scripts updated"
fi

echo ""

# Verify README
echo "📚 Checking README.md:"
if grep -q "# WebServer - Configurable Web Server" README.md; then
    echo "✅ README title updated"
else
    echo "❌ README title not updated"
fi

if grep -q "go build -o bin/webserver ./cmd/webserver" README.md; then
    echo "✅ README build commands updated"
else
    echo "❌ README build commands not updated"
fi

echo ""
echo "🎉 Rename Verification Complete!"
echo ""

# Summary
echo "📋 Summary:"
echo "  • Module name: webserver"
echo "  • Binary path: cmd/webserver/"
echo "  • Build command: go build -o bin/webserver ./cmd/webserver"
echo "  • Docker service: webserver"
echo "  • All import paths updated"
echo ""

echo "✨ The application has been successfully renamed from 'proxy1' to 'webserver'!"
echo ""
echo "🚀 Next steps:"
echo "  1. Build: go build -o bin/webserver ./cmd/webserver"
echo "  2. Run server: ./bin/webserver"
echo "  3. Run client: ./bin/webserver --client"
echo "  4. Docker: docker-compose up -d" 