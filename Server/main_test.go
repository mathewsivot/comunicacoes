package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/go/crypto/datatypes"
	"example.com/go/crypto/protocol"
	"github.com/gorilla/websocket"
)

func TestWebSocketHandshakeAndRateFlow(t *testing.T) {
	for _, operationMode := range []string{protocol.OperationModeGoBackN, protocol.OperationModeSelectiveRepeat} {
		t.Run(operationMode, func(t *testing.T) {
			server := newSocketServer(func(currency string) (*datatypes.Rate, error) {
				return &datatypes.Rate{Currency: strings.ToUpper(currency), Price: 123.45}, nil
			})

			testServer := httptest.NewServer(httpHandler(server))
			defer testServer.Close()

			wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + websocketPath
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("dial websocket: %v", err)
			}
			defer conn.Close()

			err = conn.WriteJSON(protocol.HandshakeRequest{
				Type:           protocol.MessageTypeHandshakeRequest,
				OperationMode:  operationMode,
				MaxMessageSize: 512,
			})
			if err != nil {
				t.Fatalf("write handshake: %v", err)
			}

			var handshake protocol.HandshakeResponse
			if err := conn.ReadJSON(&handshake); err != nil {
				t.Fatalf("read handshake: %v", err)
			}

			if handshake.Status != protocol.HandshakeStatusAccepted {
				t.Fatalf("unexpected handshake status: %s", handshake.Status)
			}

			if handshake.OperationMode != operationMode {
				t.Fatalf("unexpected operation mode: %s", handshake.OperationMode)
			}

			if handshake.MaxMessageSize != 512 {
				t.Fatalf("unexpected negotiated max size: %d", handshake.MaxMessageSize)
			}

			err = conn.WriteJSON(protocol.RateRequest{
				Type:     protocol.MessageTypeRateRequest,
				Currency: "btc",
			})
			if err != nil {
				t.Fatalf("write rate request: %v", err)
			}

			var response protocol.RateResponse
			if err := conn.ReadJSON(&response); err != nil {
				t.Fatalf("read rate response: %v", err)
			}

			if response.Currency != "BTC" {
				t.Fatalf("unexpected currency: %s", response.Currency)
			}

			if response.Price != 123.45 {
				t.Fatalf("unexpected price: %f", response.Price)
			}
		})
	}
}

func TestWebSocketRejectsRateRequestBeforeHandshake(t *testing.T) {
	server := newSocketServer(func(currency string) (*datatypes.Rate, error) {
		return &datatypes.Rate{Currency: currency, Price: 1}, nil
	})

	testServer := httptest.NewServer(httpHandler(server))
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + websocketPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	err = conn.WriteJSON(protocol.RateRequest{
		Type:     protocol.MessageTypeRateRequest,
		Currency: "BTC",
	})
	if err != nil {
		t.Fatalf("write rate request: %v", err)
	}

	var response protocol.ErrorResponse
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("read error response: %v", err)
	}

	if response.Type != protocol.MessageTypeError {
		t.Fatalf("unexpected response type: %s", response.Type)
	}
}

func TestWebSocketRejectsUnsupportedOperationMode(t *testing.T) {
	server := newSocketServer(func(currency string) (*datatypes.Rate, error) {
		return &datatypes.Rate{Currency: currency, Price: 1}, nil
	})

	testServer := httptest.NewServer(httpHandler(server))
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + websocketPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	err = conn.WriteJSON(protocol.HandshakeRequest{
		Type:           protocol.MessageTypeHandshakeRequest,
		OperationMode:  "invalid",
		MaxMessageSize: 512,
	})
	if err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	var response protocol.ErrorResponse
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("read error response: %v", err)
	}

	if response.Type != protocol.MessageTypeError {
		t.Fatalf("unexpected response type: %s", response.Type)
	}
}

func httpHandler(server *socketServer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(websocketPath, server.handleWebSocket)
	return mux
}
