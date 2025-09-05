#!/bin/bash

echo "=== Running Currency Exchange Service Tests ==="
echo

echo "1. Running Unit Tests..."
go test ./internal/service -v

echo
echo "2. Running Benchmark Tests..."
go test ./internal/service -bench=. -benchmem

echo
echo "3. Running Integration Tests (requires server to be stopped)..."
echo "To run integration tests with a live server, use:"
INTEGRATION=1 go test -run TestIntegrationOnly -v

echo
echo "4. Code Coverage..."
go test ./internal/service -cover

echo
echo "=== Test Summary Complete ==="
