package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/router"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/ilya372317/must-have-metrics/internal/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var sLogger = logger.Get()

func main() {
	if err := run(); err != nil {
		sLogger.Panicf("failed to start server: %v", err)
	}
}

func run() error {
	repository := storage.NewInMemoryStorage()
	if err := godotenv.Load(utils.Root + "/.env-server"); err != nil {
		logger.Get().Warnf("failed load .env-server file: %v", err)
	}
	cnfg, err := config.NewServer()
	if err != nil {
		sLogger.Panicf("failed get server config: %v", err)
	}

	db, err := sql.Open("pgx", cnfg.DatabaseDSN)
	if err != nil {
		sLogger.Fatalf("failed init postgres database: %v", err)
	}

	if err = db.Ping(); err != nil {
		sLogger.Warn(err)
		//TODO: set repo to in memory implementation
	}

	if cnfg.DatabaseDSN != "" {
		runMigrations(db)
	}

	if cnfg.StoreInterval > 0 {
		go service.SaveDataToFilesystemByInterval(cnfg, repository)
	}
	if cnfg.Restore {
		if err = service.FillFromFilesystem(repository, cnfg.FilePath); err != nil {
			sLogger.Warn(err)
		}
	}
	sLogger.Infof("server is starting...")
	err = http.ListenAndServe(
		cnfg.Host,
		router.AlertRouter(repository, cnfg),
	)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func runMigrations(db *sql.DB) {
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		sLogger.Fatalf("failed init postgres driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+utils.Root+"/db/migrations",
		"metrics", driver)
	if err != nil {
		sLogger.Fatalf("failed get migration instance: %v", err)
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			sLogger.Fatalf("failed run migrations: %v", err)
		}
	}
}
