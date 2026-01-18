package broker

import (
	"errors"
	"queuego/pkg/types"
	"sync"
	"time"
)

type BrokerConfig struct {
	MaxQueueSize    int
	MessageTTL      time.Duration
	CleanupInterval time.Duration
}

type Broker struct {
	Topics map[string]*Topic
	Config BrokerConfig
	mu     sync.RWMutex

	// metrics
	TotalMessages       int
	ActiveSubscriptions int

	stopCleanup chan struct{}
}

// NewBroker initializes a broker with the given config.
func NewBroker(config BrokerConfig) *Broker {
	return &Broker{
		Topics:      make(map[string]*Topic),
		Config:      config,
		stopCleanup: make(chan struct{}),
	}
}

// Start begins broker operations and starts cleanup goroutine.
func (b *Broker) Start() {
	go b.cleanupLoop()
}

// Stop gracefully shuts down the broker.
func (b *Broker) Stop() {
	close(b.stopCleanup)

	b.mu.Lock()
	defer b.mu.Unlock()
	for _, topic := range b.Topics {
		topic.Close()
	}
}

// CreateTopic creates a new topic if it does not exist.
func (b *Broker) CreateTopic(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.Topics[name]; exists {
		return errors.New("topic already exists")
	}

	b.Topics[name] = NewTopic(name, b.Config.MaxQueueSize)
	return nil
}

// DeleteTopic deletes a topic and cleans up subscriptions.
func (b *Broker) DeleteTopic(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if topic, exists := b.Topics[name]; exists {
		topic.Close()
		delete(b.Topics, name)
		return nil
	}
	return errors.New("topic not found")
}

// GetTopic returns a topic by name.
func (b *Broker) GetTopic(name string) (*Topic, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if topic, exists := b.Topics[name]; exists {
		return topic, nil
	}
	return nil, errors.New("topic not found")
}

// Publish adds a message to a topic, creating the topic if necessary.
func (b *Broker) Publish(topicName string, msg *types.Message) error {
	b.mu.Lock()
	topic, exists := b.Topics[topicName]
	if !exists {
		topic = NewTopic(topicName, b.Config.MaxQueueSize)
		b.Topics[topicName] = topic
	}
	b.mu.Unlock()

	if err := topic.Publish(msg); err != nil {
		return err
	}

	b.mu.Lock()
	b.TotalMessages++
	b.mu.Unlock()
	return nil
}

// Subscribe adds a subscriber to a topic.
func (b *Broker) Subscribe(topicName, clientID string) (*Subscription, error) {
	b.mu.Lock()
	topic, exists := b.Topics[topicName]
	if !exists {
		topic = NewTopic(topicName, b.Config.MaxQueueSize)
		b.Topics[topicName] = topic
	}
	b.mu.Unlock()

	sub := NewSubscription(clientID+"-"+topicName, topicName, clientID, 100, nil)
	topic.AddSubscription(sub)

	b.mu.Lock()
	b.ActiveSubscriptions++
	b.mu.Unlock()
	return sub, nil
}

// Unsubscribe removes a subscription by ID.
func (b *Broker) Unsubscribe(subID string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, topic := range b.Topics {
		topic.RemoveSubscription(subID)
	}
}

// cleanupLoop periodically removes expired messages.
func (b *Broker) cleanupLoop() {
	ticker := time.NewTicker(b.Config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.RLock()
			for _, topic := range b.Topics {
				topic.Queue.RemoveExpired(b.Config.MessageTTL)
			}
			b.mu.RUnlock()
		case <-b.stopCleanup:
			return
		}
	}
}
