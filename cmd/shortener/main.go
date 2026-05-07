package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var urlStore = make(map[string]string)

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	id := generateID()
	urlStore[id] = string(body)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + id))
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

func newRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", shortenHandler)
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
	return http.ListenAndServe(":8080", newRouter())
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
