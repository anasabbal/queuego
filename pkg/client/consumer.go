package client

import (
	"queuego/internal/protocol"
	"time"
)

type Consumer struct {
	*Client
	bufferSze int
}

func NewConsumer(cfg ClientConfig) *Consumer {
	return &Consumer{
		Client: &Client{
			Config: cfg,
			active: false,
		},
	}
}
func (c *Consumer) Subscribe(topic string, handler func(msg *protocol.Command)) error {
	cmd := &protocol.Command{
		Type:  protocol.SUBSCRIBE,
		Topic: topic,
	}

	if err := c.SendCommand(cmd); err != nil {
		return err
	}

	// track subscriber
	c.mu.Lock()
	if c.subscribers == nil {
		c.subscribers = make(map[string]func(*protocol.Command))
	}
	c.subscribers[topic] = handler
	c.mu.Unlock()

	// start goroutine to read messages
	go c.readLoop(topic, handler)
	return nil
}

func (c *Consumer) readLoop(topic string, handler func(msg *protocol.Command)) {
	for c.active {
		msg, err := c.ReadResponse()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if msg.Topic != topic {
			continue
		}

		handler(msg)

		// send ACK
		ack := &protocol.Command{
			Type:      protocol.SUBSCRIBE,
			MessageID: msg.MessageID,
			Topic:     topic,
		}
		_ = c.SendCommand(ack)
	}
}

func (c *Consumer) Unsubscribe(topic string) error {
	cmd := &protocol.Command{
		Type:  protocol.UNSUBSCRIBE,
		Topic: topic,
	}

	if err := c.SendCommand(cmd); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribers, topic)
	return nil
}
