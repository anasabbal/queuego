package types

import (
	"errors"
	"time"
)

// message represents a generic message structure.
type Message struct {
	ID        string            // UUID or unique identifier
	Topic     string            // message topic
	Payload   []byte            // raw message payload
	Timestamp time.Time         // message creation time
	Headers   map[string]string // optional metadata
	Priority  int               // message priority (higher = more priority)
}

// NewMessage creates a new Message with the current timestamp.
func NewMessage(id, topic string, payload []byte, headers map[string]string, priority int) *Message {
	return &Message{
		ID:        id,
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
		Headers:   headers,
		Priority:  priority,
	}
}

// validate checks whether the message is valid
func (m *Message) Validate() error {
	if m == nil {
		return errors.New("message is nil")
	}
	if m.ID == "" {
		return errors.New("message ID is required")
	}
	if m.Topic == "" {
		return errors.New("message topic is required")
	}
	if m.Payload == nil {
		return errors.New("message payload is required")
	}
	if m.Timestamp.IsZero() {
		return errors.New("message timestamp is required")
	}
	if m.Priority < 0 {
		return errors.New("message priority must be non-negative")
	}
	return nil
}
