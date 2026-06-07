package repository

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type URLRepository interface {
	Save(id, url string)
	Get(id string) (string, bool)
	GenerateID() string
}

type InMemoryRepository struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		data: make(map[string]string),
	}
}

func (r *InMemoryRepository) Save(id, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[id] = url
}

func (r *InMemoryRepository) Get(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.data[id]
	return url, ok
}

func (r *InMemoryRepository) GenerateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
