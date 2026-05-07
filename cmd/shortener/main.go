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

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if r.URL.Path != "/" {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil || len(body) == 0 {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			id := generateID()
			urlStore[id] = string(body)

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("http://localhost:8080/" + id))

		case http.MethodGet:
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

		default:
			http.Error(w, "", http.StatusBadRequest)
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
