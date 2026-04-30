package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"example.com/go/crypto/protocol"
	"github.com/gorilla/websocket"
)

const (
	serverURL            = "ws://localhost:3000/ws"
	minMessageSize int64 = 30
)

func main() {
	inputReader := bufio.NewReader(os.Stdin)
	operationMode := promptOperationMode(inputReader)
	maxMessageSize := promptMaxMessageSize(inputReader)

	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	defer conn.Close()

	// HANDSHAKE
	err = conn.WriteJSON(protocol.HandshakeRequest{
		Type:           protocol.MessageTypeHandshakeRequest,
		OperationMode:  operationMode,
		MaxMessageSize: maxMessageSize,
	})
	if err != nil {
		fmt.Println("Erro ao enviar handshake:", err)
		return
	}

	var handshake protocol.HandshakeResponse
	err = conn.ReadJSON(&handshake)
	if err != nil {
		fmt.Println("Erro ao ler handshake:", err)
		return
	}

	if handshake.Type != protocol.MessageTypeHandshakeResponse ||
		handshake.Status != protocol.HandshakeStatusAccepted {
		fmt.Println("Servidor recusou a conexao:", handshake.Error)
		return
	}

	fmt.Printf(
		"Handshake concluido. Modo: %s | Tamanho maximo: %d bytes\n",
		handshake.OperationMode,
		handshake.MaxMessageSize,
	)

	// LOOP DO CLIENTE
	for {
		fmt.Print("\nDigite o codigo da moeda (ex: BTC, ETH) ou 'exit': ")
		currency, _ := inputReader.ReadString('\n')
		currency = strings.TrimSpace(currency)

		if strings.ToLower(currency) == "exit" {
			fmt.Println("Encerrando cliente...")
			break
		}

		err := conn.WriteJSON(protocol.RateRequest{
			Type:     protocol.MessageTypeRateRequest,
			Currency: currency,
		})
		if err != nil {
			fmt.Println("Erro ao enviar requisicao:", err)
			break
		}

		var response protocol.RateResponse
		err = conn.ReadJSON(&response)
		if err != nil {
			fmt.Println("Erro ao ler resposta:", err)
			break
		}

		if response.Error != "" {
			fmt.Println("Resposta do Servidor:", response.Error)
			continue
		}

		fmt.Printf(
			"Resposta do Servidor: Currency:%s | Price:%.2f\n",
			response.Currency,
			response.Price,
		)
	}
}


func promptOperationMode(inputReader *bufio.Reader) string {
	for {
		fmt.Print("Escolha o modo de operacao (gbn/sr): ")
		mode, _ := inputReader.ReadString('\n')
		mode = strings.ToLower(strings.TrimSpace(mode))

		if mode == protocol.OperationModeGoBackN ||
			mode == protocol.OperationModeSelectiveRepeat {
			return mode
		}

		fmt.Println("Modo invalido. Use 'gbn' ou 'sr'.")
	}
}

func promptMaxMessageSize(inputReader *bufio.Reader) int64 {
	for {
		fmt.Print("Escolha o max_message_size da sessao em bytes: ")
		value, _ := inputReader.ReadString('\n')
		value = strings.TrimSpace(value)

		maxMessageSize, err := strconv.ParseInt(value, 10, 64)
		if err == nil && maxMessageSize >= minMessageSize {
			return maxMessageSize
		}

		fmt.Printf(
			"Valor invalido. Informe um numero inteiro maior ou igual a %d.\n",
			minMessageSize,
		)
	}
}