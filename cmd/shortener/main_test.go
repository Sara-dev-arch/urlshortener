package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		bodyCheck   func(t *testing.T, body string)
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test #1",
			body: "https://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
				bodyCheck: func(t *testing.T, body string) {
					assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"), "shortened URL should have base prefix")
				},
			},
		},
		{
			name: "empty body",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				bodyCheck:   nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlStore = make(map[string]string)

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			w := httptest.NewRecorder()
			shortenHandler(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			if test.want.bodyCheck != nil {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				test.want.bodyCheck(t, string(resBody))
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name      string
		storeID   string
		storeURL  string
		requestID string
		want      want
	}{
		{
			name:      "positive test #1",
			storeID:   "EwHXdJfB",
			storeURL:  "https://practicum.yandex.ru/",
			requestID: "EwHXdJfB",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/",
			},
		},
		{
			name:      "id not found",
			storeID:   "EwHXdJfB",
			storeURL:  "https://practicum.yandex.ru/",
			requestID: "notfound",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlStore = make(map[string]string)
			urlStore[test.storeID] = test.storeURL

			request := httptest.NewRequest(http.MethodGet, "/"+test.requestID, nil)
			w := httptest.NewRecorder()
			redirectHandler(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}

func TestShortenHandlerMethodNotAllowed(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "GET / returns 400",
			method: http.MethodGet,
		},
		{
			name:   "DELETE / returns 400",
			method: http.MethodDelete,
		},
		{
			name:   "PUT / returns 400",
			method: http.MethodPut,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", nil)
			w := httptest.NewRecorder()
			shortenHandler(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))
		})
	}
}
