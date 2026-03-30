package protocol

const (
	MessageTypeHandshakeRequest  = "handshake_request"
	MessageTypeHandshakeResponse = "handshake_response"
	MessageTypeRateRequest       = "rate_request"
	MessageTypeRateResponse      = "rate_response"
	MessageTypeError             = "error"

	OperationModeGoBackN         = "gbn"
	OperationModeSelectiveRepeat = "sr"

	HandshakeStatusAccepted = "accepted"
)

type HandshakeRequest struct {
	Type           string `json:"type"`
	OperationMode  string `json:"operation_mode"`
	MaxMessageSize int64  `json:"max_message_size"`
}

type HandshakeResponse struct {
	Type           string `json:"type"`
	Status         string `json:"status"`
	OperationMode  string `json:"operation_mode"`
	MaxMessageSize int64  `json:"max_message_size"`
	Error          string `json:"error,omitempty"`
}

type RateRequest struct {
	Type     string `json:"type"`
	Currency string `json:"currency"`
}

type RateResponse struct {
	Type     string  `json:"type"`
	Currency string  `json:"currency,omitempty"`
	Price    float64 `json:"price,omitempty"`
	Error    string  `json:"error,omitempty"`
}

type ErrorResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}
