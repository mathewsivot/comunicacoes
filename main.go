package main

import (
	"fmt"
	"sync"

	"example.com/go/crypto/api"
)

func main() {
	currencies := []string{"BTC", "ETH", "BCH", "ADA"}
	var wg sync.WaitGroup
	for _, currency := range currencies {
		wg.Add(1)
		go func(currencyCode string) {
			getCurrencyData(currencyCode)
			wg.Done()
		}(currency)

	}
	wg.Wait()
}

func getCurrencyData(currency string) {
	rates, err := api.GetRate(currency)
	if err == nil {
		fmt.Printf("%v : %v \n", rates.Currency, rates.Price)
	}
}
