package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/router"
	"github.com/ilya372317/must-have-metrics/internal/storage"
)

const defaultServerAddress = "localhost:8080"
const staticFilePath = "static"

var (
	repository storage.Storage
	host       *string
)

func init() {
	repository = storage.MakeInMemoryStorage()
	cnfg := new(config.ServerConfig)
	if err := cnfg.Init(); err != nil {
		log.Fatalln(err.Error())
	}
	host = flag.String("a", defaultServerAddress, "server address")

	if cnfg.Host != "" {
		host = &cnfg.Host
	}
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatalf("failed to start server on address %s: %v", *host, err)
	}
}

func run() error {
	return http.ListenAndServe(*host, router.AlertRouter(repository, staticFilePath))
}
