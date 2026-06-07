package main

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Sara-dev-arch/urlshortener/internal/config"
	"github.com/Sara-dev-arch/urlshortener/internal/handler"
	"github.com/Sara-dev-arch/urlshortener/internal/logger"
	"github.com/Sara-dev-arch/urlshortener/internal/repository"
	"github.com/Sara-dev-arch/urlshortener/internal/service"
)

func run() error {
	if err := logger.Initialize("info"); err != nil {
		return err
	}

	cfg := config.Parse()

	var repo repository.URLRepository
	if cfg.FileStoragePath != "" {
		var err error
		repo, err = repository.NewPersistentRepository(cfg.FileStoragePath)
		if err != nil {
			return err
		}
	} else {
		repo = repository.NewInMemoryRepository()
	}

	svc := service.NewURLService(repo, cfg.BaseURL)
	h := handler.NewURLHandler(svc)
	r := handler.NewRouter(h)

	logger.Log.Info("Running server", zap.String("address", cfg.ServerAddress))
	return http.ListenAndServe(cfg.ServerAddress, r)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
