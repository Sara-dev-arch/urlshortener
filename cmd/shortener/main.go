package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Sara-dev-arch/urlshortener/internal/config"
	"github.com/go-chi/chi/v5"
)

var urlStore = make(map[string]string)

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func shortenHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		id := generateID()
		urlStore[id] = string(body)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(baseURL + "/" + id))
	}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

func apiShortenHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req shortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		id := generateID()
		urlStore[id] = req.URL

		resp := shortenResponse{Result: baseURL + "/" + id}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	originalURL, ok := urlStore[id]
	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func newRouter(baseURL string) chi.Router {
	r := chi.NewRouter()
	r.Post("/", shortenHandler(baseURL))
	r.Post("/api/shorten", apiShortenHandler(baseURL))
	r.Get("/{id}", redirectHandler)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusBadRequest)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusBadRequest)
	})
	return r
}

func run() error {
	cfg := config.Parse()
	return http.ListenAndServe(cfg.ServerAddress, newRouter(cfg.BaseURL))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
