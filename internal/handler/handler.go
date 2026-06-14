package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Sara-dev-arch/urlshortener/internal/logger"
	"github.com/Sara-dev-arch/urlshortener/internal/middleware"
	"github.com/Sara-dev-arch/urlshortener/internal/model"
	"github.com/Sara-dev-arch/urlshortener/internal/service"
	"github.com/go-chi/chi/v5"
)

type URLHandler struct {
	svc *service.URLService
}

func NewURLHandler(svc *service.URLService) *URLHandler {
	return &URLHandler{svc: svc}
}

func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	result := h.svc.Shorten(string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result))
}

func (h *URLHandler) APIShorten(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	result := h.svc.Shorten(req.URL)
	resp := model.ShortenResponse{Result: result}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	originalURL, ok := h.svc.Expand(id)
	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func NewRouter(h *URLHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Gzip)
	r.Use(logger.RequestLogger)
	r.Post("/", h.Shorten)
	r.Post("/api/shorten", h.APIShorten)
	r.Get("/{id}", h.Redirect)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusBadRequest)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusBadRequest)
	})
	return r
}
