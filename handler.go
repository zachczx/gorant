package main

import (
	"log"
	"net/http"
	"time"
)

// type Log struct {
// 	handler http.Handler
// }

// func (l *Log) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	l.handler.ServeHTTP(w, r)
// }

// func TimerLogger(handler http.Handler) *Log {
// 	return &Log{handler}
// }

// Logger

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

		log.Printf(">>>>[%v] --- Route[%v] --- [%v] --- [%v]\r\n\r\n", rec.status, r.URL, since, r.Proto)
	})
}
