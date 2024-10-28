package main

import (
	"fmt"
	"net/http"
	"time"
)

type Compressor struct {
	handler http.Handler
}

func (h Compressor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.handler.ServeHTTP(w, r)
	since := time.Since(start)
	fmt.Println(since)
}

func Sandwicher(hd http.Handler) *Compressor {
	return &Compressor{hd}
}
