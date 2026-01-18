package queue

import (
	"errors"
	"queuego/pkg/types"
	"sync"
)

// queue represents a thread-safe message queue.
type Queue struct {
	mu       sync.Mutex
	messages []*types.Message
	MaxSize  int
}

// new queue creates a new queue with optional max size 0
func NewQueue(maxSize int) *Queue {
	return &Queue{
		messages: []*types.Message{},
		MaxSize:  maxSize,
	}
}

// push adds a message to the end of the queue
func (q *Queue) Push(msg *types.Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.MaxSize > 0 && len(q.messages) >= q.MaxSize {
		return errors.New("queue is full")
	}
	q.messages = append(q.messages, msg)
	return nil
}

// pop removes and returns the message from the front of the queue.
func (q *Queue) Pop() (*types.Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) == 0 {
		return nil, errors.New("queue is empty")
	}

	msg := q.messages[0]
	q.messages = q.messages[1:]
	return msg, nil
}

// peek returns the message at the front without removing it.
func (q *Queue) Peek() (*types.Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) == 0 {
		return nil, errors.New("queue is empty")
	}

	return q.messages[0], nil
}

// len returns the current number of messages in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.messages)
}

// clear removes all messages from the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = []*types.Message{}
}
