package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"currency_go_microservice/internal/service"
)

// newLocalClient creates an HTTP client that bypasses any proxy settings.
func newLocalClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Disable proxies entirely to avoid corporate proxy interference with localhost
	transport.Proxy = nil
	return &http.Client{Transport: transport, Timeout: 5 * time.Second}
}

// waitForHealthy polls /health until the service reports healthy or timeout.
func waitForHealthy(t *testing.T, client *http.Client, baseURL string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp != nil {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var health map[string]interface{}
				if json.Unmarshal(body, &health) == nil {
					if v, ok := health["status"]; ok {
						switch vv := v.(type) {
						case string:
							if strings.EqualFold(vv, "healthy") {
								return
							}
						case bool:
							if vv {
								return
							}
						}
					}
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("service did not become healthy at %s within %s", baseURL, timeout)
}

// Integration Tests
func TestCurrencyExchangeServiceIntegration(t *testing.T) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}

	client := newLocalClient()

	// Optionally start a local server when running outside CI
	var cmd *exec.Cmd
	if os.Getenv("START_LOCAL") == "1" {
		cmd = exec.Command("go", "run", "cmd/main.go")
		cmd.Dir = "."
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start local server: %v", err)
		}
		defer func() {
			if cmd.Process != nil {
				_ = cmd.Process.Signal(syscall.SIGTERM)
				_, _ = cmd.Process.Wait()
			}
		}()
	}

	// Wait for the service to be ready
	waitForHealthy(t, client, baseURL, 10*time.Second)

	t.Run("Health Check Integration", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Failed to make health check request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var healthResp map[string]interface{}
		if err := json.Unmarshal(body, &healthResp); err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}

		// Accept either {"status":"healthy"} or {"status":true}
		if v, ok := healthResp["status"]; ok {
			okStatus := false
			switch vv := v.(type) {
			case string:
				okStatus = strings.EqualFold(vv, "healthy")
			case bool:
				okStatus = vv
			}
			if !okStatus {
				t.Errorf("Expected status healthy/true, got %#v", v)
			}
		} else {
			t.Errorf("Missing 'status' field in health response")
		}
	})

	t.Run("Rates Endpoint Integration", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/rates")
		if err != nil {
			t.Fatalf("Failed to make rates request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var ratesResp map[string]interface{}
		err = json.Unmarshal(body, &ratesResp)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}

		if ratesResp["base"] != "USD" {
			t.Errorf("Expected base currency 'USD', got '%s'", ratesResp["base"])
		}

		rates, ok := ratesResp["rates"].(map[string]interface{})
		if !ok {
			t.Errorf("Expected rates to be a map")
		}

		// Verify some known currencies exist
		expectedCurrencies := []string{"USD", "EUR", "GBP", "JPY"}
		for _, currency := range expectedCurrencies {
			if _, exists := rates[currency]; !exists {
				t.Errorf("Expected currency %s to exist in rates", currency)
			}
		}
	})

	t.Run("Currency Exchange Integration", func(t *testing.T) {
		testCases := []struct {
			name           string
			from           string
			to             string
			amount         string
			expectedStatus int
		}{
			{
				name:           "Valid USD to EUR conversion",
				from:           "USD",
				to:             "EUR",
				amount:         "100",
				expectedStatus: http.StatusOK,
			},
			{
				name:           "Valid EUR to GBP conversion",
				from:           "EUR",
				to:             "GBP",
				amount:         "50",
				expectedStatus: http.StatusOK,
			},
			{
				name:           "Invalid currency conversion",
				from:           "XYZ",
				to:             "EUR",
				amount:         "100",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "Invalid amount",
				from:           "USD",
				to:             "EUR",
				amount:         "invalid",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := fmt.Sprintf("%s/exchange?from=%s&to=%s&amount=%s", baseURL, tc.from, tc.to, tc.amount)
				resp, err := client.Get(url)
				if err != nil {
					t.Fatalf("Failed to make exchange request: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != tc.expectedStatus {
					t.Errorf("Expected status code %d, got %d", tc.expectedStatus, resp.StatusCode)
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				if tc.expectedStatus == http.StatusOK {
					var exchangeResp service.ExchangeResponse
					err = json.Unmarshal(body, &exchangeResp)
					if err != nil {
						t.Errorf("Failed to parse JSON response: %v", err)
					}

					if exchangeResp.From != tc.from {
						t.Errorf("Expected from currency %s, got %s", tc.from, exchangeResp.From)
					}

					if exchangeResp.To != tc.to {
						t.Errorf("Expected to currency %s, got %s", tc.to, exchangeResp.To)
					}

					if exchangeResp.ConvertedAmount <= 0 {
						t.Errorf("Expected positive converted amount, got %f", exchangeResp.ConvertedAmount)
					}

					if exchangeResp.Rate <= 0 {
						t.Errorf("Expected positive exchange rate, got %f", exchangeResp.Rate)
					}
				} else {
					var errorResp service.ErrorResponse
					err = json.Unmarshal(body, &errorResp)
					if err != nil {
						t.Errorf("Failed to parse error response: %v", err)
					}

					if errorResp.Error == "" {
						t.Errorf("Expected error message but got empty string")
					}
				}
			})
		}
	})

	t.Run("Content Type Headers Integration", func(t *testing.T) {
		endpoints := []string{"/health", "/rates", "/exchange?from=USD&to=EUR&amount=100"}

		for _, endpoint := range endpoints {
			t.Run(endpoint, func(t *testing.T) {
				resp, err := client.Get(baseURL + endpoint)
				if err != nil {
					t.Fatalf("Failed to make request to %s: %v", endpoint, err)
				}
				defer resp.Body.Close()

				contentType := resp.Header.Get("Content-Type")
				if !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
					t.Errorf("Expected Content-Type starting with 'application/json', got '%s'", contentType)
				}
			})
		}
	})

	t.Run("Method Not Allowed Integration", func(t *testing.T) {
		// Test POST method on exchange endpoint
		resp, err := client.Post(baseURL+"/exchange", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make POST request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var errorResp service.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		if err != nil {
			t.Errorf("Failed to parse error response: %v", err)
		}

		if errorResp.Error != "Only GET method is allowed" {
			t.Errorf("Expected specific error message, got '%s'", errorResp.Error)
		}
	})
}

// Benchmark tests for performance using service directly
func BenchmarkConvertCurrency(b *testing.B) {
	cs := service.NewCurrencyService()
	for i := 0; i < b.N; i++ {
		cs.ConvertCurrency("USD", "EUR", 100.0)
	}
}

func BenchmarkExchangeHandler(b *testing.B) {
	cs := service.NewCurrencyService()
	req, _ := http.NewRequest("GET", "/exchange?from=USD&to=EUR&amount=100", nil)

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(cs.ExchangeHandler)
		handler.ServeHTTP(rr, req)
	}
}

// Test that runs only when INTEGRATION environment variable is set
func TestIntegrationOnly(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION=1 to run.")
	}
	TestCurrencyExchangeServiceIntegration(t)
}
