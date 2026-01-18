package broker

import (
	"queuego/pkg/types"
	"sync"
	"time"
)

type MessageState int

const (
	PENDING MessageState = iota
	DELIVERED
	ACKNOWLEDGED
)

type BrokerMessage struct {
	Msg           *types.Message
	DeliveryCount int
	ExpiresAt     time.Time
	AckDeadline   time.Time
	State         MessageState
	mu            sync.Mutex
}

// NewBrokerMessage creates a new broker message with optional TTL
func NewBrokerMessage(msg *types.Message, ttl time.Duration) *BrokerMessage {
	bm := &BrokerMessage{
		Msg:           msg,
		DeliveryCount: 0,
		State:         PENDING,
	}
	if ttl > 0 {
		bm.ExpiresAt = time.Now().Add(ttl)
	}
	return bm
}
func (bm *BrokerMessage) MarkDelivered() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.State = DELIVERED
	bm.DeliveryCount++
	bm.AckDeadline = time.Now().Add(30 * time.Second)
}

// ack marks the message as acknowledged.
func (bm *BrokerMessage) Ack() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.State = ACKNOWLEDGED
}

// isExpired checks if the message has expired.
func (bm *BrokerMessage) IsExpired() bool {
	if bm.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(bm.ExpiresAt)
}
