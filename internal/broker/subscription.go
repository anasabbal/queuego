package broker

import (
	"errors"
	"queuego/pkg/types"
	"time"
)

type Subscription struct {
	ID             string
	Topic          string
	ClientID       string
	MessageChannel chan *types.Message
	Active         bool
	Filter         func(*types.Message) bool
}

// NewSubscription creates a new subscription with a buffered channel.
func NewSubscription(id, topic, clientID string, buffer int, filter func(*types.Message) bool) *Subscription {
	return &Subscription{
		ID:             id,
		Topic:          topic,
		ClientID:       clientID,
		MessageChannel: make(chan *types.Message, buffer),
		Active:         true,
		Filter:         filter,
	}
}

// send pushes a message to the subscriber channel with non-blocking logic.
func (s *Subscription) Send(msg *types.Message, timeout time.Duration) error {
	if !s.Active {
		return errors.New("subscription inactive")
	}

	if s.Filter != nil && !s.Filter(msg) {
		return nil
	}

	select {
	case s.MessageChannel <- msg:
		return nil
	case <-time.After(timeout):
		return errors.New("send to subscription timed out")
	}
}

// close closes the subscription and cleans up resources.
func (s *Subscription) Close() {
	s.Active = false
	close(s.MessageChannel)
}
