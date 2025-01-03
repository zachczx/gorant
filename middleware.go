package main

import (
	"compress/gzip"
	"fmt"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/go-swiss/compress"
	"github.com/pterm/pterm"
)

// Taken from Go-Chi package
var defaultCompressibleContentTypes = []string{
	"text/html",
	"text/css",
	"text/plain",
	"text/javascript",
	"application/javascript",
	"application/x-javascript",
	"application/json",
	"application/atom+xml",
	"application/rss+xml",
	"image/svg+xml",
}

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
		// Eventstream doesn't seem to have header? So it causes panic when WriteHeader is called.
		if strings.Contains(r.URL.Path, "/event") {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := StatusRecorder{w, 200}
		next.ServeHTTP(&rec, r)
		since := time.Since(start)

		if !strings.Contains(r.URL.Path, "/static/") {
			var status string
			var method string
			var url string
			var duration string
			var protocol string

			switch rec.status {
			case 200:
				status = pterm.Green(strconv.Itoa(rec.status))
			default:
				status = pterm.Red(strconv.Itoa(rec.status))
			}

			switch r.Method {
			case "GET":
				method = pterm.Green(r.Method)
			case "POST":
				method = pterm.Blue(r.Method)
			default:
				method = r.Method
			}

			switch r.URL.String() {
			case "":
				url = pterm.Red(r.URL)
			default:
				url = r.URL.String()
			}

			switch {
			case since < time.Millisecond*100:
				duration = pterm.Green(since)
			default:
				duration = pterm.Red(since)
			}

			switch w.Header().Get("Content-Encoding") {
			case "br":
				protocol = pterm.LightWhite(w.Header().Get("Content-Encoding"))
			default:
				protocol = pterm.Red(w.Header().Get("Content-Encoding"))
			}
			fmt.Println(" ")
			pterm.DefaultSection.Println("Request!")
			pterm.Printf("[%v]-[%v]-[%v]-[%v]-[%v]\r\n\r\n", status, method, protocol, duration, url)
			fmt.Println("###################")
		}
	})
}

// Implements Gzip for SSE compression with added flushing.
type ZResponseWriter struct {
	http.ResponseWriter
	GzipWriter *gzip.Writer
}

func (zrw *ZResponseWriter) Write(data []byte) (int, error) {
	n, err := zrw.GzipWriter.Write(data)
	if err != nil {
		return n, err
	}
	err = zrw.GzipWriter.Flush()
	return n, err
}

// This is needed, else SSE doesn't stream.
func (zrw *ZResponseWriter) Flush() {
	zrw.GzipWriter.Flush()
	if flusher, ok := zrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Implements Brotli for SSE compression with added flushing.
type ZBrResponseWriter struct {
	http.ResponseWriter
	BrWriter *brotli.Writer
}

func (zrw *ZBrResponseWriter) Write(data []byte) (int, error) {
	n, err := zrw.BrWriter.Write(data)
	if err != nil {
		return n, err
	}
	err = zrw.BrWriter.Flush()
	return n, err
}

func (zrw *ZBrResponseWriter) Flush() {
	zrw.BrWriter.Flush()
	if flusher, ok := zrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func ZxCompress() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ae := r.Header.Get("Accept-Encoding")
		if !strings.Contains(ae, "br") && !strings.Contains(ae, "gzip") {
			fmt.Println("NOOOOOOO!!!!!!!!!!!!!!!!!!!!")
			sseHandler(w, r)
			return
		}

		if strings.Contains(ae, "br") {
			fmt.Println("BR!!!!!!")
			w.Header().Set("Content-Encoding", "br")
			brotliWriter := brotli.NewWriterLevel(w, 2)
			defer brotliWriter.Close()
			zipped := &ZBrResponseWriter{ResponseWriter: w, BrWriter: brotliWriter}
			sseHandler(zipped, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gzw := gzip.NewWriter(w)
		zipped := &ZResponseWriter{ResponseWriter: w, GzipWriter: gzw}
		defer gzw.Close()
		sseHandler(zipped, r)
	})
}

func ExcludeCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Using strings.Contain because css file has an additional charset=utf-8, which doesn't allow == match.
		// .woff2 also doesn't have a mime type. Using this because mime.TypeByExtension guesses based on extension anyway.

		ext := mime.TypeByExtension(path.Ext(r.RequestURI))

		for _, v := range defaultCompressibleContentTypes {
			if strings.Contains(ext, v) {
				ch := compress.Middleware(next)
				ch.ServeHTTP(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)

		// Not using this because it's more work.
		// if strings.Contains(r.URL.Path, "/event") {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }
		// ext := path.Ext(r.RequestURI)
		// switch ext {
		// case ".webp", ".woff2":
		// 	next.ServeHTTP(w, r)
		// default:
		// 	ch := compress.Middleware(next)
		// 	ch.ServeHTTP(w, r)
		// }
	})
}

var cacheFiles = []string{"htmx-bundle.js", ".woff2", ".webp", ".svg"}

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
			w.Header().Set("Expires", time.Now().Add(time.Second*31536000).Format(http.TimeFormat))
			// http.FileServer() sets Last-Modified header, so there's no point modifying it as I tried below.
			// Anyway it's stupid to set it to time.Now() as I originally did.
			// Read: https://stackoverflow.com/questions/47033156/overriding-last-modified-header-in-http-fileserver
			// w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		}

		next.ServeHTTP(w, r)
	})
}
