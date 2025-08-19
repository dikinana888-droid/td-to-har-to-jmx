#!/bin/bash

# Script to run TCPDump to JMX Converter API

set -e

echo "🚀 TCPDump to JMX Converter API"
echo "================================"

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Creating from example..."
    cp .env.example .env
    echo "✅ Created .env file. Please update it with your AWS credentials."
fi

# Parse command
case "$1" in
    build)
        echo "🔨 Building application..."
        go build -o tcpdump-to-jmx .
        echo "✅ Build complete"
        ;;
    
    run)
        echo "🏃 Running application..."
        if [ ! -f tcpdump-to-jmx ]; then
            echo "Binary not found. Building first..."
            go build -o tcpdump-to-jmx .
        fi
        ./tcpdump-to-jmx
        ;;
    
    docker-build)
        echo "🐳 Building Docker image..."
        docker build -t tcpdump-to-jmx:latest .
        echo "✅ Docker image built"
        ;;
    
    docker-run)
        echo "🐳 Running with Docker..."
        docker run -p 8080:8080 --env-file .env tcpdump-to-jmx:latest
        ;;
    
    compose-up)
        echo "🐳 Starting with Docker Compose..."
        docker-compose up -d
        echo "✅ Services started"
        echo "📍 API: http://localhost:8080"
        echo "📍 MinIO: http://localhost:9001 (minioadmin/minioadmin)"
        ;;
    
    compose-down)
        echo "🛑 Stopping Docker Compose services..."
        docker-compose down
        echo "✅ Services stopped"
        ;;
    
    compose-logs)
        echo "📜 Showing Docker Compose logs..."
        docker-compose logs -f
        ;;
    
    test)
        echo "🧪 Running tests..."
        go test ./... -v
        ;;
    
    clean)
        echo "🧹 Cleaning up..."
        rm -f tcpdump-to-jmx
        rm -f *.pcap *.har *.jmx
        docker-compose down -v 2>/dev/null || true
        echo "✅ Cleanup complete"
        ;;
    
    help|*)
        echo "Usage: $0 {build|run|docker-build|docker-run|compose-up|compose-down|compose-logs|test|clean|help}"
        echo ""
        echo "Commands:"
        echo "  build         - Build the Go application"
        echo "  run           - Run the application locally"
        echo "  docker-build  - Build Docker image"
        echo "  docker-run    - Run application in Docker"
        echo "  compose-up    - Start services with Docker Compose"
        echo "  compose-down  - Stop Docker Compose services"
        echo "  compose-logs  - View Docker Compose logs"
        echo "  test          - Run tests"
        echo "  clean         - Clean up build artifacts and containers"
        echo "  help          - Show this help message"
        ;;
esac