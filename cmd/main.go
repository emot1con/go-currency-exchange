package main

import (
	"fmt"
	"log"
	"net/http"

	"currency_go_microservice/internal/service"
)

func main() {
	// Create currency service instance
	currencyService := service.NewCurrencyService()

	// Set up routes
	http.HandleFunc("/exchange", currencyService.ExchangeHandler)
	http.HandleFunc("/health", currencyService.HealthHandler)
	http.HandleFunc("/rates", currencyService.RatesHandler)

	// Start server
	port := ":8080"
	fmt.Printf("Currency Exchange Service starting on port %s\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /exchange?from=USD&to=EUR&amount=100")
	fmt.Println("  GET /health")
	fmt.Println("  GET /rates")

	log.Fatal(http.ListenAndServe(port, nil))
}
