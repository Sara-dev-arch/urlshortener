package service

import "github.com/Sara-dev-arch/urlshortener/internal/repository"

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
