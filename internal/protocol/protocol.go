package protocol

type CommandType string

const (
	CONNECT     CommandType = "CONNECT"
	PUBLISH     CommandType = "PUBLISH"
	SUBSCRIBE   CommandType = "SUBSCRIBE"
	UNSUBSCRIBE CommandType = "UNSUBSCRIBE"
	ACK         CommandType = "ACK"
	PING        CommandType = "PING"
	PONG        CommandType = "PONG"
)

// status represents response status codes

type StatusCode string

const (
	OK              StatusCode = "OK"
	ERROR           StatusCode = "ERROR"
	NOT_FOUND       StatusCode = "NOT_FOUND"
	INVALID_REQUEST StatusCode = "INVALID_REQUEST"
)

// command represents a request sent from client to broker
type Command struct {
	Type      CommandType `json:"type"`
	Topic     string      `json:"topic,omitempty"`
	MessageID string      `json:"message_id,omitempty"`
	Payload   []byte      `json:"payload,omitempty"`
	Status    StatusCode
}

// response represents a broker response to a client
type Response struct {
	Status  StatusCode `json:"status"`
	Message string     `json:"message,omitempty"`
	Data    []byte     `json:"data,omitempty"`
}
