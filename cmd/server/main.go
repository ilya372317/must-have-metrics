package main

import (
	"context"
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

func main() {
	if err := logger.Init(); err != nil {
		panic(fmt.Errorf("failed init logger for server: %w", err))
	}

	if err := run(); err != nil {
		logger.Log.Panicf("failed to start server: %v", err)
	}
}

func run() error {
	var repository router.AlertStorage = storage.NewInMemoryStorage()
	if err := godotenv.Load(utils.Root + "/.env-server"); err != nil {
		logger.Log.Warnf("failed load .env-server file: %v", err)
	}
	cnfg, err := config.NewServer()
	if err != nil {
		logger.Log.Panicf("failed get server config: %v", err)
	}

	db, err := sql.Open("pgx", cnfg.DatabaseDSN)
	if err != nil {
		logger.Log.Fatalf("failed init postgres database: %v", err)
	}

	if cnfg.ShouldConnectToDatabase() {
		repository = &storage.DatabaseStorage{
			DB: db,
		}
		runMigrations(db)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cnfg.StoreInterval > 0 {
		go service.SaveDataToFilesystemByInterval(ctx, cnfg, repository)
	}
	if cnfg.Restore {
		if err = service.FillFromFilesystem(ctx, repository, cnfg.FilePath); err != nil {
			logger.Log.Warn(err)
		}
	}
	logger.Log.Infof("server is starting...")
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
		logger.Log.Fatalf("failed init postgres driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+utils.Root+"/db/migrations",
		"metrics", driver)
	if err != nil {
		logger.Log.Fatalf("failed get migration instance: %v", err)
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			logger.Log.Fatalf("failed run migrations: %v", err)
		}
	}
}
