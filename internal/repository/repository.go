package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/Sara-dev-arch/urlshortener/internal/model"
)

type URLRepository interface {
	Save(id, url string) error
	Get(id string) (string, bool)
	GetAll() []model.UserURL
}

type urlRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

func (r *InMemoryRepository) Save(id, url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[id] = url
	return nil
}

func (r *InMemoryRepository) Get(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.data[id]
	return url, ok
}

func (r *InMemoryRepository) GetAll() []model.UserURL {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.UserURL, 0, len(r.data))
	for id, originalURL := range r.data {
		result = append(result, model.UserURL{
			ShortURL:    id,
			OriginalURL: originalURL,
		})
	}
	return result
}

type PersistentRepository struct {
	mem     *InMemoryRepository
	mu      sync.Mutex
	records []urlRecord
	counter atomic.Int64
	file    *os.File
}

func NewPersistentRepository(filePath string) (*PersistentRepository, error) {
	mem := NewInMemoryRepository()
	repo := &PersistentRepository{
		mem: mem,
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read storage file: %w", err)
		}
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &repo.records); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stored records: %w", err)
		}
		for _, rec := range repo.records {
			mem.data[rec.ShortURL] = rec.OriginalURL
		}
		repo.counter.Store(int64(len(repo.records)))
	}

	repo.file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage file: %w", err)
	}

	return repo, nil
}

func (r *PersistentRepository) Save(id, url string) error {
	r.mem.mu.Lock()
	r.mem.data[id] = url
	r.mem.mu.Unlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	rec := urlRecord{
		UUID:        r.nextUUID(),
		ShortURL:    id,
		OriginalURL: url,
	}
	r.records = append(r.records, rec)

	if err := r.file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate storage file: %w", err)
	}
	if _, err := r.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek storage file: %w", err)
	}

	enc := json.NewEncoder(r.file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r.records); err != nil {
		return fmt.Errorf("failed to encode records: %w", err)
	}
	return nil
}

func (r *PersistentRepository) Get(id string) (string, bool) {
	return r.mem.Get(id)
}

func (r *PersistentRepository) GetAll() []model.UserURL {
	return r.mem.GetAll()
}

func (r *PersistentRepository) Close() error {
	return r.file.Close()
}

func (r *PersistentRepository) nextUUID() string {
	n := r.counter.Add(1)
	return strconv.FormatInt(n, 10)
}
