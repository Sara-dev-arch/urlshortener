package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

var compressibleTypes = []string{"application/json", "text/html"}

func canCompress(ct string) bool {
	for _, t := range compressibleTypes {
		if strings.HasPrefix(ct, t) {
			return true
		}
	}
	return false
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzWriter *gzip.Writer
	gzInit   bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.initGzip()
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.gzInit {
		w.initGzip()
	}
	if w.gzWriter != nil {
		return w.gzWriter.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *gzipResponseWriter) initGzip() {
	w.gzInit = true
	ct := w.Header().Get("Content-Type")
	if canCompress(ct) {
		w.gzWriter = gzip.NewWriter(w.ResponseWriter)
		w.Header().Set("Content-Encoding", "gzip")
	}
}

func (w *gzipResponseWriter) Close() {
	if w.gzWriter != nil {
		w.gzWriter.Close()
	}
}

func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = io.NopCloser(gz)
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzw := &gzipResponseWriter{ResponseWriter: w}
		defer gzw.Close()

		next.ServeHTTP(gzw, r)
	})
}
