package service

import (
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

func (s *URLService) Shorten(originalURL string) string {
	id := s.repo.GenerateID()
	s.repo.Save(id, originalURL)
	return s.baseURL + "/" + id
}

func (s *URLService) Expand(id string) (string, bool) {
	return s.repo.Get(id)
}

func (s *URLService) GetAllURLs() []model.UserURL {
	records := s.repo.GetAll()
	result := make([]model.UserURL, len(records))
	for i, r := range records {
		result[i] = model.UserURL{
			ShortURL:    s.baseURL + "/" + r.ShortURL,
			OriginalURL: r.OriginalURL,
		}
	}
	return result
}
