package storage

import (
	"bufio"
	"encoding/gob"
	"os"
	"path/filepath"
	"queuego/pkg/types"
	"sync"
)

// FileStorage implements append-only file persistence.
type FileStorage struct {
	mu      sync.RWMutex
	dir     string
	files   []string
	maxSize int64
	offsets map[string]int64 // messageID -> file offset
}

func NewFileStorage(dir string, maxSize int64) (*FileStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &FileStorage{
		dir:     dir,
		maxSize: maxSize,
		offsets: make(map[string]int64),
	}, nil
}

func (fs *FileStorage) Append(msg *types.Message) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	filePath := filepath.Join(fs.dir, "messages.log")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	if err := enc.Encode(msg); err != nil {
		return err
	}
	fs.offsets[msg.ID] = 0
	return nil
}

func (fs *FileStorage) Retrieve(msgID string) (*types.Message, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	filePath := filepath.Join(fs.dir, "messages.log")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec := gob.NewDecoder(bufio.NewReader(file))
	for {
		var msg types.Message
		if err := dec.Decode(&msg); err != nil {
			break
		}
		if msg.ID == msgID {
			return &msg, nil
		}
	}
	return nil, os.ErrNotExist
}

func (fs *FileStorage) Delete(msgID string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	delete(fs.offsets, msgID)
	// actual log compaction would be needed
	return nil
}

func (fs *FileStorage) List(topic string) ([]*types.Message, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var result []*types.Message
	filePath := filepath.Join(fs.dir, "messages.log")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec := gob.NewDecoder(bufio.NewReader(file))
	for {
		var msg types.Message
		if err := dec.Decode(&msg); err != nil {
			break
		}
		if msg.Topic == topic {
			result = append(result, &msg)
		}
	}
	return result, nil
}
