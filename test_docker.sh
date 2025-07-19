#!/bin/bash

echo "🐳 Testing Docker Setup for WebServer..."
echo ""
echo "✅ Docker Configuration Created:"
echo "  • Dockerfile - Multi-stage build with security best practices"
echo "  • .dockerignore - Optimized build context"
echo "  • docker-compose.yml - Easy deployment and development"
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    exit 1
fi

echo "🔧 Docker Setup Information:"
echo ""
echo "📦 Build Process:"
echo "  1. Build stage: golang:1.24-alpine (with dependencies)"
echo "  2. Runtime stage: alpine:latest (minimal footprint)"
echo "  3. Non-root user: webserveruser (security best practice)"
echo "  4. Health check: /stats endpoint monitoring"
echo ""

echo "🚀 Available Commands:"
echo ""
echo "┌─ Build and Run ───────────────────────────────────────────────────┐"
echo "│ # Build the Docker image                                           │"
echo "│ docker build -t webserver .                                       │"
echo "│                                                                    │"
echo "│ # Run the container                                                │"
echo "│ docker run -d -p 8080:8080 --name webserver-server webserver      │"
echo "│                                                                    │"
echo "│ # Run with custom config                                           │"
echo "│ docker run -d -p 8080:8080 \\                                      │"
echo "│   -v /path/to/config.json:/app/configs/default.json:ro \\         │"
echo "│   webserver                                                        │"
echo "└────────────────────────────────────────────────────────────────────┘"
echo ""

echo "┌─ Docker Compose ──────────────────────────────────────────────────┐"
echo "│ # Build and start the service                                      │"
echo "│ docker-compose up -d                                               │"
echo "│                                                                    │"
echo "│ # View logs                                                        │"
echo "│ docker-compose logs -f webserver                                   │"
echo "│                                                                    │"
echo "│ # Stop the service                                                 │"
echo "│ docker-compose down                                                │"
echo "│                                                                    │"
echo "│ # Development mode (with volume mounts)                           │"
echo "│ docker-compose --profile dev up webserver-dev                     │"
echo "└────────────────────────────────────────────────────────────────────┘"
echo ""

echo "🧪 Test Docker Build:"
echo ""
read -p "Do you want to test the Docker build now? (y/N): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Building Docker image..."
    if docker build -t webserver:test .; then
        echo "✅ Docker build successful!"
        echo ""
        
        read -p "Do you want to run the container for testing? (y/N): " -n 1 -r
        echo ""
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "Starting container..."
            CONTAINER_ID=$(docker run -d -p 8081:8080 --name webserver-test webserver:test)
            
            if [ $? -eq 0 ]; then
                echo "✅ Container started with ID: ${CONTAINER_ID:0:12}"
                echo ""
                
                # Wait for container to be ready
                echo "Waiting for service to start..."
                sleep 5
                
                # Test the service
                echo "Testing HTTP endpoints..."
                if curl -s http://localhost:8081/stats > /dev/null; then
                    echo "✅ Service is responding on http://localhost:8081"
                    echo ""
                    echo "🌐 Available endpoints:"
                    echo "  • http://localhost:8081/ - Static files"
                    echo "  • http://localhost:8081/stats - Server statistics"
                    echo "  • http://localhost:8081/config - Configuration"
                    echo "  • http://localhost:8081/api/delay - Test delay endpoint"
                    echo "  • http://localhost:8081/api/error - Test error endpoint"
                    echo ""
                    echo "Test the service and press ENTER when done..."
                    read -r
                else
                    echo "❌ Service is not responding properly"
                fi
                
                # Cleanup
                echo "Cleaning up test container..."
                docker stop webserver-test >/dev/null 2>&1
                docker rm webserver-test >/dev/null 2>&1
                echo "✅ Test container removed"
            else
                echo "❌ Failed to start container"
            fi
        fi
        
        # Clean up test image
        read -p "Remove test image? (y/N): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker rmi webserver:test >/dev/null 2>&1
            echo "✅ Test image removed"
        fi
    else
        echo "❌ Docker build failed"
        exit 1
    fi
fi

echo ""
echo "✨ Docker Setup Complete!"
echo ""
echo "📋 Created Files:"
echo "  ✅ Dockerfile - Multi-stage production-ready build"
echo "  ✅ .dockerignore - Optimized build context"
echo "  ✅ docker-compose.yml - Easy orchestration"
echo ""
echo "🔒 Security Features:"
echo "  • Non-root user execution"
echo "  • Minimal alpine base image"
echo "  • Health check monitoring"
echo "  • No unnecessary packages"
echo ""
echo "🚀 Production Ready:"
echo "  • Multi-stage build for smaller images"
echo "  • Proper layer caching for faster builds"
echo "  • Configuration file mounting support"
echo "  • Container orchestration with Docker Compose"
echo ""
echo "🎉 Your application is now containerized and ready for deployment!" 