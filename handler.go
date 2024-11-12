package main

import (
	"log"
	"net/http"
	"path"
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

		log.Printf(">>>>[%v] --- Route[%v] --- [%v] --- [%v] --- [%v]\r\n\r\n", rec.status, r.URL, since, r.Proto, w.Header().Get("Content-Encoding"))
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
