package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
)

var urlStore = make(map[string]string)

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

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
	id := r.URL.Path[1:]
	if id == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	originalURL, ok := urlStore[id]
	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, shortenHandler)
	mux.HandleFunc(`/{id}`, redirectHandler)

	return http.ListenAndServe(":8080", mux)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
