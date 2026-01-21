package storage

import (
	"queuego/pkg/types"
	"sync"
	"time"
)

type MemoryStorage struct {
	mu       sync.RWMutex
	messages map[string]*types.Message
	topics   map[string]map[string]struct{} // topic -> message IDs
	order    []string                       // for LRU eviction
	MaxSize  int
}

func NewMemoryStorage(maxSize int) *MemoryStorage {
	return &MemoryStorage{
		messages: make(map[string]*types.Message),
		topics:   make(map[string]map[string]struct{}),
		order:    []string{},
		MaxSize:  maxSize,
	}
}

// store adds a message to a storage
func (s *MemoryStorage) Store(msg *types.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.MaxSize > 0 && len(s.messages) >= s.MaxSize {
		s.evictLRU()
	}

	s.messages[msg.ID] = msg
	s.order = append(s.order, msg.ID)

	if _, ok := s.topics[msg.Topic]; !ok {
		s.topics[msg.Topic] = make(map[string]struct{})
	}
	s.topics[msg.Topic][msg.ID] = struct{}{}

	return nil
}

func (s *MemoryStorage) evictLRU() {
	if len(s.order) == 0 {
		return
	}
	oldestID := s.order[0]
	s.order = s.order[1:]
	if msg, ok := s.messages[oldestID]; ok {
		delete(s.messages, oldestID)
		if topicSet, ok := s.topics[msg.Topic]; ok {
			delete(topicSet, oldestID)
			if len(topicSet) == 0 {
				delete(s.topics, msg.Topic)
			}
		}
	}
}
func (s *MemoryStorage) RemoveExpired(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()

	for id, msg := range s.messages {
		if !msg.Timestamp.IsZero() && msg.Timestamp.Add(ttl).Before(now) {
			delete(s.messages, id)
			for i, oid := range s.order {
				if oid == id {
					s.order = append(s.order[:i], s.order[i+1:]...)
					break
				}
			}
			if topicSet, ok := s.topics[msg.Topic]; ok {
				delete(topicSet, id)
				if len(topicSet) == 0 {
					delete(s.topics, msg.Topic)
				}
			}
		}
	}
}
