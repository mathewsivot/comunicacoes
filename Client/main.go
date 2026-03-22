package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// 1. Conecta ao servidor
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// --- PASSO 1: Enviar Handshake ---
	fmt.Println("Enviando Handshake...")
	fmt.Fprintf(conn, "HELLO\n")

	status, _ := reader.ReadString('\n')
	if strings.TrimSpace(status) != "READY" {
		fmt.Println("Servidor recusou a conexão.")
		return
	}
	fmt.Println("Servidor pronto!")

	// --- PASSO 2: Solicitar Moeda ---
	fmt.Print("Digite o código da moeda (ex: BTC, ETH): ")
	inputReader := bufio.NewReader(os.Stdin)
	currency, _ := inputReader.ReadString('\n')

	fmt.Fprintf(conn, currency) // Envia para o servidor

	// --- PASSO 3: Receber Resposta Final ---
	result, _ := reader.ReadString('\n')
	fmt.Printf("Resposta do Servidor: %s", result)
}
