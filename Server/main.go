package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"example.com/go/crypto/api"
	"example.com/go/crypto/datatypes"
	"example.com/go/crypto/protocol"
	"github.com/gorilla/websocket"
)

const (
	serverAddress        = ":3000"
	websocketPath        = "/ws"
	minMessageSize       = int64(30)
	serverMaxMessageSize = int64(1024)
)

type rateFetcher func(currency string) (*datatypes.Rate, error)

type socketServer struct {
	upgrader websocket.Upgrader
	getRate  rateFetcher
}

func newSocketServer(fetcher rateFetcher) *socketServer {
	return &socketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		getRate: fetcher,
	}
}

func main() {
	server := newSocketServer(api.GetRate)

	http.HandleFunc(websocketPath, server.handleWebSocket)

	fmt.Printf(
		"Servidor WebSocket aguardando conexoes em %s%s\n",
		serverAddress,
		websocketPath,
	)

	log.Fatal(http.ListenAndServe(serverAddress, nil))
}

func (s *socketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("erro no upgrade websocket: %v", err)
		return
	}
	defer conn.Close()

	conn.SetReadLimit(serverMaxMessageSize)

	handshake, ok := s.readHandshake(conn)
	if !ok {
		return
	}

	effectiveMaxSize := minInt64(
		handshake.MaxMessageSize,
		serverMaxMessageSize,
	)

	conn.SetReadLimit(effectiveMaxSize)

	err = s.writeJSON(conn, protocol.HandshakeResponse{
		Type:           protocol.MessageTypeHandshakeResponse,
		Status:         protocol.HandshakeStatusAccepted,
		OperationMode:  handshake.OperationMode,
		MaxMessageSize: effectiveMaxSize,
	})
	if err != nil {
		log.Printf("erro ao enviar handshake response: %v", err)
		return
	}

	log.Printf(
		"handshake confirmado: conexao com %s | modo=%s | max_message_size=%d",
		conn.RemoteAddr(),
		handshake.OperationMode,
		effectiveMaxSize,
	)
	for {

		request, ok := s.readRateRequest(conn)
		if !ok {
			log.Printf("cliente desconectado: %s", conn.RemoteAddr())
			return
		}

		rate, err := s.getRate(request.Currency)
		if err != nil {
			_ = s.writeJSON(conn, protocol.RateResponse{
				Type:  protocol.MessageTypeRateResponse,
				Error: "Erro: Moeda nao encontrada",
			})
			continue
		}

		err = s.writeJSON(conn, protocol.RateResponse{
			Type:     protocol.MessageTypeRateResponse,
			Currency: rate.Currency,
			Price:    rate.Price,
		})

		if err != nil {
			log.Printf("erro ao enviar rate response: %v", err)
			return
		}
	}
}
func (s *socketServer) readHandshake(conn *websocket.Conn) (protocol.HandshakeRequest, bool) {

	var request protocol.HandshakeRequest

	if err := conn.ReadJSON(&request); err != nil {
		s.writeProtocolError(conn, "Erro: handshake invalido")
		return protocol.HandshakeRequest{}, false
	}

	switch {
	case request.Type != protocol.MessageTypeHandshakeRequest:
		s.writeProtocolError(conn, "Erro: primeira mensagem deve ser handshake_request")
		return protocol.HandshakeRequest{}, false

	case !isSupportedOperationMode(request.OperationMode):
		s.writeProtocolError(conn, "Erro: modo de operacao nao suportado")
		return protocol.HandshakeRequest{}, false

	case request.MaxMessageSize < minMessageSize:
		s.writeProtocolError(conn, "Erro: tamanho maximo invalido")
		return protocol.HandshakeRequest{}, false
	}

	return request, true
}

func isSupportedOperationMode(mode string) bool {
	return mode == protocol.OperationModeGoBackN ||
		mode == protocol.OperationModeSelectiveRepeat
}
func (s *socketServer) readRateRequest(conn *websocket.Conn) (protocol.RateRequest, bool) {

	var request protocol.RateRequest

	if err := conn.ReadJSON(&request); err != nil {
		s.writeProtocolError(conn, "Erro: requisicao invalida")
		return protocol.RateRequest{}, false
	}

	if request.Type != protocol.MessageTypeRateRequest {
		s.writeProtocolError(conn, "Erro: mensagem esperada rate_request")
		return protocol.RateRequest{}, false
	}

	request.Currency = strings.TrimSpace(request.Currency)

	if request.Currency == "" {
		s.writeProtocolError(conn, "Erro: moeda obrigatoria")
		return protocol.RateRequest{}, false
	}

	return request, true
}
func (s *socketServer) writeProtocolError(conn *websocket.Conn, message string) {

	_ = s.writeJSON(conn, protocol.ErrorResponse{
		Type:  protocol.MessageTypeError,
		Error: message,
	})

	_ = conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.ClosePolicyViolation, message),
	)
}

func (s *socketServer) writeJSON(conn *websocket.Conn, payload any) error {
	return conn.WriteJSON(payload)
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}