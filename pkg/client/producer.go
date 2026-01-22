package client

import (
	"fmt"
	"queuego/internal/protocol"
	"queuego/pkg/types"
)

type Producer struct {
	*Client
}

func (p *Producer) Publish(topic string, payload []byte) error {
	msg := &types.Message{
		Topic:   topic,
		Payload: payload,
	}
	cmd := &protocol.Command{
		Type:    protocol.PUBLISH,
		Topic:   topic,
		Payload: msg.Payload,
	}

	// send the command over TCP connection
	if err := p.SendCommand(cmd); err != nil {
		return err
	}

	// wait for broker response
	resp, err := p.ReadResponse()
	if err != nil {
		return err
	}
	if resp.Type != protocol.ACK {
		return fmt.Errorf("publish failed, got type %s", resp.Type)
	}
	return nil
}

// PublishBatch sends multiple messages together (simplified)
func (p *Producer) PublishBatch(topic string, payloads [][]byte) error {
	for _, payload := range payloads {
		if err := p.Publish(topic, payload); err != nil {
			return err
		}
	}
	return nil
}

// PublishAsync sends a message asynchronously with callback.
func (p *Producer) PublishAsync(topic string, payload []byte, callback func(error)) {
	go func() {
		err := p.Publish(topic, payload)
		if callback != nil {
			callback(err)
		}
	}()
}
