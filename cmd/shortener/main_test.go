package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestShortenHandler(t *testing.T) {
	srv := httptest.NewServer(newRouter("http://localhost:8080"))
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
			urlStore = make(map[string]string)

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
	srv := httptest.NewServer(newRouter("http://localhost:8080"))
	defer srv.Close()

	testCases := []struct {
		name             string
		storeID          string
		storeURL         string
		requestID        string
		expectedCode     int
		expectedLocation string
	}{
		{
			name:             "positive test #1",
			storeID:          "EwHXdJfB",
			storeURL:         "https://practicum.yandex.ru/",
			requestID:        "EwHXdJfB",
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
		{
			name:             "id not found",
			storeID:          "EwHXdJfB",
			storeURL:         "https://practicum.yandex.ru/",
			requestID:        "notfound",
			expectedCode:     http.StatusBadRequest,
			expectedLocation: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			urlStore = make(map[string]string)
			urlStore[tc.storeID] = tc.storeURL

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
