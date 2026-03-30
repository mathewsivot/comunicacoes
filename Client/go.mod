module example.com/go/client

go 1.25.3

require (
	example.com/go/crypto v0.0.0
	github.com/gorilla/websocket v1.5.3
)

replace example.com/go/crypto => ../Server
