package protocol

const (
	// MaxBodyBytes is the largest buffered HTTP request or response body carried
	// by the current JSON/base64 relay protocol.
	MaxBodyBytes = 10 * 1024 * 1024
	// MaxMessageBytes leaves room for base64 expansion and bounded HTTP headers.
	MaxMessageBytes = 16 * 1024 * 1024
)

// MessageType defines the kind of WebSocket frame exchanged between CLI and server.
type MessageType string

const (
	// Client -> Server: initial handshake
	TypeHello MessageType = "hello"

	// Server -> Client: incoming HTTP request to proxy
	TypeRequest MessageType = "req"

	// Client -> Server: HTTP response from local service
	TypeResponse MessageType = "res"

	// Server -> Client: keepalive ping
	TypePing MessageType = "ping"

	// Client -> Server: keepalive pong
	TypePong MessageType = "pong"
)

// Message is the envelope for all WebSocket traffic.
type Message struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

// Hello is sent by the CLI immediately after the WebSocket upgrade.
type Hello struct {
	Type     string `json:"type"`
	Token    string `json:"token"`
	TunnelID string `json:"tunnel_id"`
}

// ProxyRequest is sent by the server when an HTTP request arrives for a tunnel.
type ProxyRequest struct {
	Type    string            `json:"type"`
	ReqID   string            `json:"req_id"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Host    string            `json:"host"`
	Headers map[string]string `json:"headers"`
	BodyB64 string            `json:"body_b64"`
}

// ProxyResponse is sent by the CLI after it forwards the request to localhost.
type ProxyResponse struct {
	Type    string            `json:"type"`
	ReqID   string            `json:"req_id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	BodyB64 string            `json:"body_b64"`
}

// Ping is a server-side keepalive.
type Ping struct {
	Type string `json:"type"`
}

// Pong is the client keepalive reply.
type Pong struct {
	Type string `json:"type"`
}
