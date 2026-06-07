package main

import (
	"net/http"

	"github.com/Sara-dev-arch/urlshortener/internal/config"
	"github.com/Sara-dev-arch/urlshortener/internal/handler"
	"github.com/Sara-dev-arch/urlshortener/internal/repository"
	"github.com/Sara-dev-arch/urlshortener/internal/service"
)

func run() error {
	cfg := config.Parse()

	repo := repository.NewInMemoryRepository()
	svc := service.NewURLService(repo, cfg.BaseURL)
	h := handler.NewURLHandler(svc)
	r := handler.NewRouter(h)

	return http.ListenAndServe(cfg.ServerAddress, r)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
