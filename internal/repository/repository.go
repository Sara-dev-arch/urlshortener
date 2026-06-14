package repository

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/Sara-dev-arch/urlshortener/internal/model"
)

type URLRepository interface {
	Save(id, url string)
	Get(id string) (string, bool)
	GenerateID() string
	GetAll() []model.UserURL
}

type urlRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type InMemoryRepository struct {
	mu       sync.RWMutex
	data     map[string]string
	records  []urlRecord
	filePath string
	counter  atomic.Int64
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

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return repo, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return repo, nil
	}

	if err := json.Unmarshal(data, &repo.records); err != nil {
		return nil, err
	}

	for _, rec := range repo.records {
		repo.data[rec.ShortURL] = rec.OriginalURL
	}
	repo.counter.Store(int64(len(repo.records)))

	return repo, nil
}

func (r *InMemoryRepository) Save(id, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[id] = url

	if r.filePath == "" {
		return
	}

	rec := urlRecord{
		UUID:        r.nextUUID(),
		ShortURL:    id,
		OriginalURL: url,
	}
	r.records = append(r.records, rec)

	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	enc.Encode(r.records)
}

func (r *InMemoryRepository) nextUUID() string {
	n := r.counter.Add(1)
	return strconv.FormatInt(n, 10)
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

func (r *InMemoryRepository) GenerateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
