package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// ExchangeRates holds the conversion rates from USD to other currencies
var ExchangeRates = map[string]float64{
	"USD": 1.0,
	"EUR": 0.85,
	"GBP": 0.73,
	"JPY": 110.0,
	"CAD": 1.25,
	"AUD": 1.35,
	"CHF": 0.92,
	"CNY": 6.45,
	"INR": 74.5,
	"BRL": 5.2,
}

// ExchangeRequest represents the request structure
type ExchangeRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

// ExchangeResponse represents the response structure
type ExchangeResponse struct {
	From            string  `json:"from"`
	To              string  `json:"to"`
	Amount          float64 `json:"amount"`
	ConvertedAmount float64 `json:"converted_amount"`
	Rate            float64 `json:"rate"`
}

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// CurrencyService handles currency exchange operations
type CurrencyService struct{}

// NewCurrencyService creates a new currency service instance
func NewCurrencyService() *CurrencyService {
	return &CurrencyService{}
}

// ConvertCurrency performs the currency conversion
func (cs *CurrencyService) ConvertCurrency(from, to string, amount float64) (float64, float64, error) {
	fromRate, fromExists := ExchangeRates[strings.ToUpper(from)]
	toRate, toExists := ExchangeRates[strings.ToUpper(to)]

	if !fromExists {
		return 0, 0, fmt.Errorf("currency %s not supported", from)
	}
	if !toExists {
		return 0, 0, fmt.Errorf("currency %s not supported", to)
	}

	// Convert to USD first, then to target currency
	usdAmount := amount / fromRate
	convertedAmount := usdAmount * toRate
	rate := toRate / fromRate

	return convertedAmount, rate, nil
}

// ExchangeHandler handles currency exchange requests
func (cs *CurrencyService) ExchangeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Only GET method is allowed"})
		return
	}

	// Parse query parameters
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")

	if from == "" || to == "" || amountStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing required parameters: from, to, amount"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	log.Println("Parsed amount:", amount)
	log.Println("to:", to)
	log.Println("from:", from)

	if err != nil || amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid amount parameter"})
		return
	}

	convertedAmount, rate, err := cs.ConvertCurrency(from, to, amount)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	response := ExchangeResponse{
		From:            strings.ToUpper(from),
		To:              strings.ToUpper(to),
		Amount:          amount,
		ConvertedAmount: convertedAmount,
		Rate:            rate,
	}

	json.NewEncoder(w).Encode(response)
}

// HealthHandler handles health check requests
func (cs *CurrencyService) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// RatesHandler returns all available exchange rates
func (cs *CurrencyService) RatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"base":  "USD",
		"rates": ExchangeRates,
	})
}
