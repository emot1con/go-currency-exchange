# Currency Exchange Microservice

A simple HTTP-based currency exchange service built in Go that provides real-time currency conversion capabilities.

## Features

- **Currency Conversion**: Convert amounts between different currencies
- **Health Check**: Monitor service health status
- **Exchange Rates**: View all available exchange rates
- **RESTful API**: Clean HTTP endpoints with JSON responses
- **Comprehensive Testing**: Unit tests, integration tests, and benchmarks

## Supported Currencies

- USD (US Dollar) - Base currency
- EUR (Euro)
- GBP (British Pound)
- JPY (Japanese Yen)
- CAD (Canadian Dollar)
- AUD (Australian Dollar)
- CHF (Swiss Franc)
- CNY (Chinese Yuan)
- INR (Indian Rupee)
- BRL (Brazilian Real)

## API Endpoints

### GET /exchange
Convert currency amounts between different currencies.

**Parameters:**
- `from` (required): Source currency code (e.g., "USD")
- `to` (required): Target currency code (e.g., "EUR")
- `amount` (required): Amount to convert (positive number)

**Example:**
```bash
curl "http://localhost:8080/exchange?from=USD&to=EUR&amount=100"
```

**Response:**
```json
{
  "from": "USD",
  "to": "EUR",
  "amount": 100,
  "converted_amount": 85,
  "rate": 0.85
}
```

### GET /health
Check service health status.

**Example:**
```bash
curl "http://localhost:8080/health"
```

**Response:**
```json
{
  "status": "healthy"
}
```

### GET /rates
Get all available exchange rates.

**Example:**
```bash
curl "http://localhost:8080/rates"
```

**Response:**
```json
{
  "base": "USD",
  "rates": {
    "USD": 1.0,
    "EUR": 0.85,
    "GBP": 0.73,
    "JPY": 110.0,
    ...
  }
}
```

## Project Structure

```
currency-go-microservice/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   └── service/
│       ├── currency.go            # Core service logic
│       └── currency_test.go       # Unit tests
├── integration_test.go            # Integration tests
├── run_tests.sh                   # Test runner script
├── go.mod                         # Go module file
└── README.md                      # This file
```

## Getting Started

### Prerequisites
- Go 1.21 or higher

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd currency-go-microservice
```

2. Initialize Go modules:
```bash
go mod tidy
```

3. Run the service:
```bash
go run cmd/main.go
```

The service will start on `http://localhost:8080`

### Testing

#### Run All Tests
```bash
./run_tests.sh
```

#### Unit Tests Only
```bash
go test ./internal/service -v
```

#### Benchmark Tests
```bash
go test ./internal/service -bench=.
```

#### Integration Tests
To run integration tests with a live server:
```bash
INTEGRATION=1 go test -run TestIntegrationOnly -v
```

#### Code Coverage
```bash
go test ./internal/service -cover
```

## Performance

The service is designed for high performance:

- **Currency Conversion**: ~21 ns/op (no allocations)
- **HTTP Handler**: ~1970 ns/op (25 allocations)
- **Code Coverage**: 100%

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- `200 OK`: Success
- `400 Bad Request`: Invalid parameters or unsupported currency
- `405 Method Not Allowed`: Invalid HTTP method

Error responses follow this format:
```json
{
  "error": "Error description"
}
```

## Architecture

The service follows clean architecture principles:

- **Service Layer** (`internal/service`): Contains business logic and HTTP handlers
- **Main Package** (`cmd/main.go`): Application entry point and server setup
- **Separation of Concerns**: Business logic is separated from HTTP handling

## Development

### Adding New Currencies

1. Update the `ExchangeRates` map in `internal/service/currency.go`
2. Add corresponding test cases in `internal/service/currency_test.go`
3. Run tests to ensure everything works

### Adding New Endpoints

1. Add handler method to `CurrencyService` struct
2. Register the handler in `cmd/main.go`
3. Add corresponding tests

## Docker Support

To run with Docker:

```bash
# Build image
docker build -t currency-exchange .

# Run container
docker run -p 8080:8080 currency-exchange
```

### Testing Docker Setup

Use the provided test script to verify Docker setup locally:

```bash
./test_docker.sh
```

This script will:
- Build a fresh Docker image
- Start the container
- Run health checks
- Execute integration tests
- Clean up resources

## CI/CD Pipeline

The project includes a Jenkins pipeline (`Jenkinsfile`) that runs:

1. **Checkout**: Retrieves source code and verifies Go installation
2. **Tests** (parallel execution):
   - Unit Tests: Core functionality testing
   - Benchmark Tests: Performance measurements
   - Integration Tests: End-to-end testing with Docker
   - Coverage: Code coverage analysis
3. **Build Binary**: Compiles the Go application
4. **Build Docker Image**: Creates containerized version
5. **Push Docker Image**: Publishes to Docker registry

### Jenkins Pipeline Features

- **Fresh Docker builds**: Each test run uses a newly built image
- **Robust health checks**: Waits for service readiness before testing
- **Comprehensive logging**: Container logs and detailed error reporting
- **Automatic cleanup**: Ensures resources are cleaned up even on failure
- **Parallel execution**: Multiple test types run simultaneously for efficiency

### Troubleshooting Integration Tests

If integration tests fail:

1. Check container logs: `docker logs currency-exchange-test`
2. Verify port availability: `netstat -tlnp | grep 8080`
3. Test health endpoint manually: `curl http://localhost:8080/health`
4. Run the local test script: `./test_docker.sh`

Common issues:
- **Port conflicts**: Ensure port 8080 is available
- **Container startup time**: Service needs time to initialize
- **Network connectivity**: Verify localhost/127.0.0.1 accessibility

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License.
