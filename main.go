package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	//"sync"

	"example.com/go/crypto/api"
)

func main() {
	http.HandleFunc("/currency", currencyHandler)

	fmt.Println("Server starting on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func currencyHandler(w http.ResponseWriter, r *http.Request) {
	currencyCode := r.URL.Query().Get("code")

	if currencyCode == "" {
		http.Error(w, "no code guiven", http.StatusBadRequest)
		return
	}
	rate, err := api.GetRate(strings.ToUpper(currencyCode))
	if err != nil {
		http.Error(w, fmt.Sprintf("Data Fetch Error: %v", err), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rate)
}

/*func main() {
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
*/
