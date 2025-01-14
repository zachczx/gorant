package main

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/go-swiss/compress"
	"github.com/pterm/pterm"
)

var defaultCompressibleFileExtensions = []string{".html", ".htm", ".css", ".txt", ".js", ".map", ".json", ".xml", ".svg"}

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
		rec := StatusRecorder{w, http.StatusOK}
		next.ServeHTTP(&rec, r)
		since := time.Since(start)

		if !strings.Contains(r.URL.Path, "/static/") {
			var status, method, url, duration, protocol string

			switch rec.status {
			case http.StatusOK:
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
		return n, fmt.Errorf("error: gzipwriter write: %w", err)
	}
	if err = zrw.GzipWriter.Flush(); err != nil {
		return n, fmt.Errorf("error: gzipwriter flush: %w", err)
	}
	return n, nil
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
		return n, fmt.Errorf("error: brotli writer write: %w", err)
	}
	if err := zrw.BrWriter.Flush(); err != nil {
		return n, fmt.Errorf("error: brotli writer flush: %w", err)
	}
	return n, nil
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
		// Templ live reload doesn't seem to work with Brotli compression, so skip compression if it's a dev env.
		if os.Getenv("DEV_ENV") == "TRUE" {
			next.ServeHTTP(w, r)
			return
		}
		// Not using mime.TypeByExtension because it doesn't have .woff2 as a mime type so it's messy
		// to handle this and html from file paths.
		ext := path.Ext(r.RequestURI)
		// Check to see if it's a router path, if so it'll be a HTML response so let's just compress it.
		if ext == "" {
			ch := compress.Middleware(next)
			ch.ServeHTTP(w, r)
			return
		}
		for _, v := range defaultCompressibleFileExtensions {
			if ext == v {
				ch := compress.Middleware(next)
				ch.ServeHTTP(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
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
