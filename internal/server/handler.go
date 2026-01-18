package server

import (
	"queuego/internal/broker"
	"queuego/internal/protocol"
	"queuego/pkg/types"
)

type Handler struct {
	Broker *broker.Broker
}

// HandleCommand routes commands to broker operations
func (h *Handler) HandleCommand(conn *Connection, cmd *protocol.Command) {
	switch cmd.Type {
	case protocol.PUBLISH:
		msg := &types.Message{
			ID:      cmd.MessageID,
			Topic:   cmd.Topic,
			Payload: cmd.Payload,
		}
		h.Broker.Publish(cmd.Topic, msg)
	case protocol.SUBSCRIBE:
		sub, _ := h.Broker.Subscribe(cmd.Topic, conn.ClientID)
		conn.Subscriptions[sub.ID] = true
	case protocol.UNSUBSCRIBE:
		h.Broker.Unsubscribe(cmd.Topic)
	case protocol.PING:
		conn.Send(&protocol.Command{Type: protocol.PONG})
	}
}
