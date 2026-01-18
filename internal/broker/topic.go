package broker

import (
	"queuego/internal/queue"
	"queuego/pkg/types"
	"sync"
	"time"
)

type Topic struct {
	Name          string
	Queue         *queue.Queue
	Subscriptions map[string]*Subscription
	mu            sync.RWMutex

	MessageCount    int
	SubscriberCount int
	LastPublish     time.Time

	stopChan chan struct{}
}

// NewTopic creates a new topic.
func NewTopic(name string, maxQueueSize int) *Topic {
	t := &Topic{
		Name:          name,
		Queue:         queue.NewQueue(maxQueueSize),
		Subscriptions: make(map[string]*Subscription),
		stopChan:      make(chan struct{}),
	}
	go t.distribute()
	return t
}

// AddSubscription adds a subscriber to the topic.
func (t *Topic) AddSubscription(sub *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Subscriptions[sub.ID] = sub
	t.SubscriberCount = len(t.Subscriptions)
}

// RemoveSubscription removes a subscriber from the topic.
func (t *Topic) RemoveSubscription(subID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if sub, ok := t.Subscriptions[subID]; ok {
		sub.Close()
		delete(t.Subscriptions, subID)
	}
	t.SubscriberCount = len(t.Subscriptions)
}

// publish adds a message to the queue.
func (t *Topic) Publish(msg *types.Message) error {
	if err := t.Queue.Push(msg); err != nil {
		return err
	}
	t.mu.Lock()
	t.MessageCount++
	t.LastPublish = time.Now()
	t.mu.Unlock()
	return nil
}

// distribute continuously reads from queue and pushes to subscribers.
func (t *Topic) distribute() {
	for {
		select {
		case <-t.stopChan:
			return
		default:
			msg, err := t.Queue.Pop()
			if err != nil {
				time.Sleep(10 * time.Millisecond) // avoid busy loop
				continue
			}

			t.mu.RLock()
			for _, sub := range t.Subscriptions {
				_ = sub.Send(msg, 50*time.Millisecond) // ignore send errors for slow subscribers
			}
			t.mu.RUnlock()
		}
	}
}

// close stops topic distribution and cleans up subscriptions.
func (t *Topic) Close() {
	close(t.stopChan)

	t.mu.Lock()
	defer t.mu.Unlock()
	for _, sub := range t.Subscriptions {
		sub.Close()
	}
	t.Subscriptions = nil
}
