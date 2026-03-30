package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"example.com/go/crypto/protocol"
	"github.com/gorilla/websocket"
)

const (
	serverURL            = "ws://localhost:3000/ws"
	clientMaxMessageSize = int64(512)
)

func main() {
	operationMode := promptOperationMode()

	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	defer conn.Close()

	if err := conn.WriteJSON(protocol.HandshakeRequest{
		Type:           protocol.MessageTypeHandshakeRequest,
		OperationMode:  operationMode,
		MaxMessageSize: clientMaxMessageSize,
	}); err != nil {
		fmt.Println("Erro ao enviar handshake:", err)
		return
	}

	var handshake protocol.HandshakeResponse
	if err := conn.ReadJSON(&handshake); err != nil {
		fmt.Println("Erro ao ler handshake:", err)
		return
	}

	if handshake.Type != protocol.MessageTypeHandshakeResponse || handshake.Status != protocol.HandshakeStatusAccepted {
		fmt.Println("Servidor recusou a conexao:", handshake.Error)
		return
	}

	if handshake.OperationMode != operationMode {
		fmt.Println("Servidor negociou um modo de operacao inesperado:", handshake.OperationMode)
		return
	}

	if handshake.MaxMessageSize > clientMaxMessageSize {
		fmt.Println("Servidor negociou um tamanho maximo invalido:", handshake.MaxMessageSize)
		return
	}

	fmt.Printf("Handshake concluido. Modo: %s | Tamanho maximo: %d bytes\n", handshake.OperationMode, handshake.MaxMessageSize)

	fmt.Print("Digite o codigo da moeda (ex: BTC, ETH): ")
	inputReader := bufio.NewReader(os.Stdin)
	currency, _ := inputReader.ReadString('\n')

	if err := conn.WriteJSON(protocol.RateRequest{
		Type:     protocol.MessageTypeRateRequest,
		Currency: strings.TrimSpace(currency),
	}); err != nil {
		fmt.Println("Erro ao enviar requisicao:", err)
		return
	}

	var response protocol.RateResponse
	if err := conn.ReadJSON(&response); err != nil {
		fmt.Println("Erro ao ler resposta:", err)
		return
	}

	if response.Error != "" {
		fmt.Println("Resposta do Servidor:", response.Error)
		return
	}

	fmt.Printf("Resposta do Servidor: Currency:%s | Price:%.2f\n", response.Currency, response.Price)
}

func promptOperationMode() string {
	inputReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Escolha o modo de operacao (gbn/sr): ")
		mode, _ := inputReader.ReadString('\n')
		mode = strings.ToLower(strings.TrimSpace(mode))

		if mode == protocol.OperationModeGoBackN || mode == protocol.OperationModeSelectiveRepeat {
			return mode
		}

		fmt.Println("Modo invalido. Use 'gbn' ou 'sr'.")
	}
}
