package main

import (
	"github.com/ilya372317/must-have-metrics/internal/utils/http/middleware"
	"io"
	"log"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("vailed to start server on port 8080: %v", err)
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "page not found")
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "url is work")
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", defaultHandler)
	mux.HandleFunc("/update/", middleware.Chain(updateHandler, middleware.ValidUpdate()))
	return http.ListenAndServe(":8080", mux)
}
