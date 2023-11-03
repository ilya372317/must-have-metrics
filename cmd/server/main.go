package main

import (
	"flag"
	"github.com/ilya372317/must-have-metrics/internal/constant"
	"github.com/ilya372317/must-have-metrics/internal/router"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"log"
	"net/http"
)

var (
	repository storage.AlertStorage
	host       *string
)

func init() {
	repository = storage.MakeInMemoryStorage()
	host = flag.String("a", ":8080", "define server address")
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatalf("vailed to start server on port 8080: %v", err)
	}
}

func run() error {
	return http.ListenAndServe(*host, router.AlertRouter(repository, constant.StaticFilePath))
}
