package server

import (
	"log"
	"queuego/internal/broker"
	"queuego/internal/protocol"
	"queuego/pkg/types"
	"time"
)

type Handler struct {
	Broker *broker.Broker
}

func (h *Handler) HandleCommand(conn *Connection, cmd *protocol.Command) {
	switch cmd.Type {

	case protocol.PUBLISH:
		msg := &types.Message{
			Topic:     cmd.Topic,
			Payload:   cmd.Payload,
			Timestamp: time.Now(),
		}

		if err := h.Broker.Publish(cmd.Topic, msg); err != nil {
			log.Printf("[%s] publish error: %v", conn.ID, err)
			conn.Send(&protocol.Command{
				Type:      protocol.ACK,
				MessageID: cmd.MessageID,
				Topic:     cmd.Topic,
				Payload:   []byte(err.Error()),
			})
			return
		}

		conn.Send(&protocol.Command{
			Type:      protocol.ACK,
			MessageID: cmd.MessageID,
			Topic:     cmd.Topic,
		})
		log.Printf("[%s] ACK sent for PUBLISH topic %s", conn.ID, cmd.Topic)

	case protocol.SUBSCRIBE:
		_, err := h.Broker.Subscribe(cmd.Topic, conn.ID)
		if err != nil {
			log.Printf("[%s] subscribe error: %v", conn.ID, err)
			conn.Send(&protocol.Command{
				Type:    protocol.ACK,
				Topic:   cmd.Topic,
				Payload: []byte(err.Error()),
			})
			return
		}

		conn.Send(&protocol.Command{
			Type:  protocol.ACK,
			Topic: cmd.Topic,
		})
		log.Printf("[%s] ACK sent for SUBSCRIBE topic %s", conn.ID, cmd.Topic)

	case protocol.UNSUBSCRIBE:
		h.Broker.Unsubscribe(conn.ID + "-" + cmd.Topic)
		conn.Send(&protocol.Command{
			Type:  protocol.ACK,
			Topic: cmd.Topic,
		})
		log.Printf("[%s] ACK sent for UNSUBSCRIBE topic %s", conn.ID, cmd.Topic)

	case protocol.PING:
		conn.Send(&protocol.Command{
			Type: protocol.PONG,
		})
		log.Printf("[%s] PONG sent", conn.ID)
	}
}
