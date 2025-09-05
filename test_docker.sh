#!/bin/bash

echo "=== Testing Docker Setup Locally ==="

# Clean up any existing container
echo "Cleaning up existing containers..."
docker stop currency-exchange-test 2>/dev/null || true
docker rm currency-exchange-test 2>/dev/null || true

# Build the image
echo "Building Docker image..."
docker build -t currency-exchange-test:latest .

if [ $? -ne 0 ]; then
    echo "❌ Docker build failed"
    exit 1
fi

# Run the container
echo "Starting container..."
docker run -d --name currency-exchange-test -p 8080:8080 currency-exchange-test:latest

if [ $? -ne 0 ]; then
    echo "❌ Failed to start container"
    exit 1
fi

# Wait for service to start
echo "Waiting for service to start..."
sleep 5

# Check if container is running
echo "Checking container status..."
docker ps | grep currency-exchange-test

# Check container logs
echo "=== Container Logs ==="
docker logs currency-exchange-test

# Test health endpoint
echo "=== Testing Health Endpoint ==="
curl -f http://localhost:8080/health || echo "Health check failed"

# Test rates endpoint
echo "=== Testing Rates Endpoint ==="
curl -f http://localhost:8080/rates || echo "Rates check failed"

# Test exchange endpoint
echo "=== Testing Exchange Endpoint ==="
curl -f "http://localhost:8080/exchange?from=USD&to=EUR&amount=100" || echo "Exchange check failed"

# Run integration tests
echo "=== Running Integration Tests ==="
INTEGRATION=1 go test -run TestIntegrationOnly -v

# Cleanup
echo "=== Cleaning up ==="
docker stop currency-exchange-test
docker rm currency-exchange-test

echo "=== Test completed ==="
