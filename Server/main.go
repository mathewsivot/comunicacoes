package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	/*
		"encoding/json"
		"log"
		"net/http"
		"sync"
	*/
	"example.com/go/crypto/api"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Falha ao iniciar servidor", err)
		return
	}
	defer ln.Close()
	fmt.Println("Servidor Socket aguardando conexao")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexao", err)
			continue
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	message, _ := reader.ReadString('\n')
	if strings.TrimSpace(message) != "HELLO" {
		conn.Write([]byte("Erro: failed handshake"))
		return
	}
	conn.Write([]byte("READY\n"))

	currency, _ := reader.ReadString('\n')
	currency = strings.TrimSpace(currency)

	rate, err := api.GetRate(currency)
	if err != nil {
		conn.Write([]byte("Erro: Moeda nao encontrada"))
		return
	}
	response := fmt.Sprintf("Currency:%s | Price:%.2f\n", rate.Currency, rate.Price)
	conn.Write([]byte(response))
}

// API format, testing serve of parsed data  -- delete/useless

/* func main() {
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
*/

// ROLLBACK TO THIS !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
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
