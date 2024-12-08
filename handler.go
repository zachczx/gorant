package main

import (
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-swiss/compress"
)

type StatusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *StatusRecorder) WriteHeader(statusCode int) {
	rec.ResponseWriter.WriteHeader(statusCode)
	rec.status = statusCode
}

func StatusLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := StatusRecorder{w, 200}
		next.ServeHTTP(&rec, r)
		since := time.Since(start)

		if !strings.Contains(r.URL.Path, "/static/") {
			log.Printf(">>>>[%v %v] -- [%v] -- [%v] -- [%v] -- [%v]\r\n\r\n", rec.status, r.Method, r.URL, since, r.Proto, w.Header().Get("Content-Encoding"))
		}
	})
}

func ExcludeCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := path.Ext(r.RequestURI)
		switch ext {
		case ".webp", ".woff2":
			h.ServeHTTP(w, r)
		default:
			ch := compress.Middleware(h)
			ch.ServeHTTP(w, r)
		}
	})
}

var cacheFiles = []string{"htmx-bundle.js", "InterVariable.woff2", "avatar-shiba.webp"}

func contains(s string, a []string) bool {
	for _, v := range a {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}

func SetCacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contains(r.URL.Path, cacheFiles) {
			// Set cache headers
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			w.Header().Set("Expires", time.Now().Add(time.Hour).Format(http.TimeFormat))
			w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		}

		next.ServeHTTP(w, r)
	})
}
