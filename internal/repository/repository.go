package repository

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"sync"
)

type URLRepository interface {
	Save(id, url string)
	Get(id string) (string, bool)
	GenerateID() string
}

type InMemoryRepository struct {
	mu       sync.RWMutex
	data     map[string]string
	filePath string
	file     *os.File
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		data: make(map[string]string),
	}
}

func NewPersistentRepository(filePath string) (*InMemoryRepository, error) {
	repo := &InMemoryRepository{
		data:     make(map[string]string),
		filePath: filePath,
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	repo.file = f

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			repo.data[parts[0]] = parts[1]
		}
	}

	return repo, nil
}

func (r *InMemoryRepository) Save(id, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[id] = url
	if r.file != nil {
		r.file.WriteString(id + " " + url + "\n")
	}
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
