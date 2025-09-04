package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Unit Tests for ConvertCurrency function
func TestConvertCurrency(t *testing.T) {
	cs := NewCurrencyService()

	tests := []struct {
		name           string
		from           string
		to             string
		amount         float64
		expectedAmount float64
		expectedRate   float64
		expectError    bool
	}{
		{
			name:           "USD to EUR conversion",
			from:           "USD",
			to:             "EUR",
			amount:         100.0,
			expectedAmount: 85.0,
			expectedRate:   0.85,
			expectError:    false,
		},
		{
			name:           "EUR to USD conversion",
			from:           "EUR",
			to:             "USD",
			amount:         85.0,
			expectedAmount: 100.0,
			expectedRate:   1.1764705882352942, // 1/0.85
			expectError:    false,
		},
		{
			name:           "Same currency conversion",
			from:           "USD",
			to:             "USD",
			amount:         100.0,
			expectedAmount: 100.0,
			expectedRate:   1.0,
			expectError:    false,
		},
		{
			name:           "Case insensitive conversion",
			from:           "usd",
			to:             "eur",
			amount:         100.0,
			expectedAmount: 85.0,
			expectedRate:   0.85,
			expectError:    false,
		},
		{
			name:        "Unsupported from currency",
			from:        "XYZ",
			to:          "USD",
			amount:      100.0,
			expectError: true,
		},
		{
			name:        "Unsupported to currency",
			from:        "USD",
			to:          "XYZ",
			amount:      100.0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertedAmount, rate, err := cs.ConvertCurrency(tt.from, tt.to, tt.amount)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if convertedAmount != tt.expectedAmount {
				t.Errorf("Expected converted amount %.2f, got %.2f", tt.expectedAmount, convertedAmount)
			}

			if rate != tt.expectedRate {
				t.Errorf("Expected rate %.10f, got %.10f", tt.expectedRate, rate)
			}
		})
	}
}

// Unit Tests for HTTP Handlers
func TestExchangeHandler(t *testing.T) {
	cs := NewCurrencyService()

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedFrom   string
		expectedTo     string
		expectedAmount float64
		expectError    bool
	}{
		{
			name:           "Valid exchange request",
			method:         "GET",
			url:            "/exchange?from=USD&to=EUR&amount=100",
			expectedStatus: http.StatusOK,
			expectedFrom:   "USD",
			expectedTo:     "EUR",
			expectedAmount: 100.0,
			expectError:    false,
		},
		{
			name:           "Missing from parameter",
			method:         "GET",
			url:            "/exchange?to=EUR&amount=100",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Missing to parameter",
			method:         "GET",
			url:            "/exchange?from=USD&amount=100",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Missing amount parameter",
			method:         "GET",
			url:            "/exchange?from=USD&to=EUR",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Invalid amount parameter",
			method:         "GET",
			url:            "/exchange?from=USD&to=EUR&amount=invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Negative amount",
			method:         "GET",
			url:            "/exchange?from=USD&to=EUR&amount=-100",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Zero amount",
			method:         "GET",
			url:            "/exchange?from=USD&to=EUR&amount=0",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Unsupported currency",
			method:         "GET",
			url:            "/exchange?from=XYZ&to=EUR&amount=100",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "POST method not allowed",
			method:         "POST",
			url:            "/exchange?from=USD&to=EUR&amount=100",
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(cs.ExchangeHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.expectError {
				var errorResp ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
				if err != nil {
					t.Errorf("Expected error response but couldn't parse JSON: %v", err)
				}
				if errorResp.Error == "" {
					t.Errorf("Expected error message but got empty string")
				}
			} else {
				var exchangeResp ExchangeResponse
				err := json.Unmarshal(rr.Body.Bytes(), &exchangeResp)
				if err != nil {
					t.Errorf("Expected exchange response but couldn't parse JSON: %v", err)
				}

				if exchangeResp.From != tt.expectedFrom {
					t.Errorf("Expected from currency %s, got %s", tt.expectedFrom, exchangeResp.From)
				}

				if exchangeResp.To != tt.expectedTo {
					t.Errorf("Expected to currency %s, got %s", tt.expectedTo, exchangeResp.To)
				}

				if exchangeResp.Amount != tt.expectedAmount {
					t.Errorf("Expected amount %.2f, got %.2f", tt.expectedAmount, exchangeResp.Amount)
				}
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	cs := NewCurrencyService()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cs.HealthHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Could not parse JSON response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}

func TestRatesHandler(t *testing.T) {
	cs := NewCurrencyService()

	req, err := http.NewRequest("GET", "/rates", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cs.RatesHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Could not parse JSON response: %v", err)
	}

	if response["base"] != "USD" {
		t.Errorf("Expected base currency 'USD', got '%s'", response["base"])
	}

	rates, ok := response["rates"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected rates to be a map")
	}

	if len(rates) != len(ExchangeRates) {
		t.Errorf("Expected %d rates, got %d", len(ExchangeRates), len(rates))
	}

	// Check if USD rate is 1.0
	if rates["USD"] != 1.0 {
		t.Errorf("Expected USD rate to be 1.0, got %v", rates["USD"])
	}
}

// Benchmark tests for performance
func BenchmarkConvertCurrency(b *testing.B) {
	cs := NewCurrencyService()
	for i := 0; i < b.N; i++ {
		cs.ConvertCurrency("USD", "EUR", 100.0)
	}
}

func BenchmarkExchangeHandler(b *testing.B) {
	cs := NewCurrencyService()
	req, _ := http.NewRequest("GET", "/exchange?from=USD&to=EUR&amount=100", nil)

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(cs.ExchangeHandler)
		handler.ServeHTTP(rr, req)
	}
}
