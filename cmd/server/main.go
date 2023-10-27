package main

import (
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/repository"
	"github.com/ilya372317/must-have-metrics/internal/service"
	"github.com/ilya372317/must-have-metrics/internal/utils/http/middleware"
	"log"
	"net/http"
)

var repoPtr repository.AlertStorage

func init() {
	repoValue := repository.MakeAlertInMemoryStorage()
	repoPtr = &repoValue
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("vailed to start server on port 8080: %v", err)
	}
}

func defaultHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "incorrect route", http.StatusBadRequest)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	updateAlertDTO := dto.CreateAlertDTOFromRequest(r)
	if err := service.AddAlert(repoPtr, updateAlertDTO); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", defaultHandler)
	mux.HandleFunc(
		"/update/",
		middleware.Chain(updateHandler, middleware.ValidUpdate(), middleware.Method(http.MethodPost)),
	)
	return http.ListenAndServe(":8080", mux)
}
