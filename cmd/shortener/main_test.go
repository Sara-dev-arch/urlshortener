package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sara-dev-arch/urlshortener/internal/handler"
	"github.com/Sara-dev-arch/urlshortener/internal/repository"
	"github.com/Sara-dev-arch/urlshortener/internal/service"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestServer() *httptest.Server {
	repo := repository.NewInMemoryRepository()
	svc := service.NewURLService(repo, "http://localhost:8080")
	h := handler.NewURLHandler(svc)
	return httptest.NewServer(handler.NewRouter(h))
}

func TestShortenHandler(t *testing.T) {
	srv := setupTestServer()
	defer srv.Close()

	successBodyCheck := func(t *testing.T, body string) {
		assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"), "shortened URL should have base prefix")
	}

	testCases := []struct {
		name         string
		method       string
		body         string
		expectedCode int
		bodyCheck    func(t *testing.T, body string)
	}{
		{
			name:         "positive test #1",
			method:       http.MethodPost,
			body:         "https://practicum.yandex.ru/",
			expectedCode: http.StatusCreated,
			bodyCheck:    successBodyCheck,
		},
		{
			name:         "empty body",
			method:       http.MethodPost,
			body:         "",
			expectedCode: http.StatusBadRequest,
			bodyCheck:    nil,
		},
		{
			name:         "method not allowed GET",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusBadRequest,
			bodyCheck:    nil,
		},
		{
			name:         "method not allowed DELETE",
			method:       http.MethodDelete,
			body:         "",
			expectedCode: http.StatusBadRequest,
			bodyCheck:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := resty.New()
			client.GetClient().CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			req := client.R()
			req.Method = tc.method
			req.URL = srv.URL + "/"
			req.SetBody(tc.body)

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")

			if tc.bodyCheck != nil {
				tc.bodyCheck(t, string(resp.Body()))
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	svc := service.NewURLService(repo, "http://localhost:8080")
	h := handler.NewURLHandler(svc)
	srv := httptest.NewServer(handler.NewRouter(h))
	defer srv.Close()

	repo.Save("EwHXdJfB", "https://practicum.yandex.ru/")

	testCases := []struct {
		name             string
		requestID        string
		expectedCode     int
		expectedLocation string
	}{
		{
			name:             "positive test #1",
			requestID:        "EwHXdJfB",
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
		{
			name:             "id not found",
			requestID:        "notfound",
			expectedCode:     http.StatusBadRequest,
			expectedLocation: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := resty.New()
			client.GetClient().CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			req := client.R()
			req.Method = http.MethodGet
			req.URL = srv.URL + "/" + tc.requestID

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tc.expectedLocation, resp.Header().Get("Location"), "Location header didn't match expected")
		})
	}
}
