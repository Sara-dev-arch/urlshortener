package service

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"

	"github.com/Sara-dev-arch/urlshortener/internal/model"
	"github.com/Sara-dev-arch/urlshortener/internal/repository"
)

type URLService struct {
	repo    repository.URLRepository
	baseURL string
}

func NewURLService(repo repository.URLRepository, baseURL string) *URLService {
	return &URLService{
		repo:    repo,
		baseURL: baseURL,
	}
}

func (s *URLService) Shorten(originalURL string) (string, error) {
	id := generateID()
	if err := s.repo.Save(id, originalURL); err != nil {
		return "", err
	}
	result, _ := url.JoinPath(s.baseURL, id)
	return result, nil
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (s *URLService) Expand(id string) (string, bool) {
	return s.repo.Get(id)
}

func (s *URLService) GetAllURLs() []model.UserURL {
	records := s.repo.GetAll()
	result := make([]model.UserURL, len(records))
	for i, r := range records {
		shortURL, _ := url.JoinPath(s.baseURL, r.ShortURL)
		result[i] = model.UserURL{
			ShortURL:    shortURL,
			OriginalURL: r.OriginalURL,
		}
	}
	return result
}
