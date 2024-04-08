// Application for collecting metrics sent by clients
//
// It is capable of storing data both in memory and in a database.
// Save its state to a file and restore it.
// It also supports multiple routes for receiving metrics in string or JSON format.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/ilya372317/must-have-metrics/internal/config"
	mygrpc "github.com/ilya372317/must-have-metrics/internal/handlers/grpc"
	http2 "github.com/ilya372317/must-have-metrics/internal/handlers/http"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/service"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/ilya372317/must-have-metrics/internal/utils"
	pb "github.com/ilya372317/must-have-metrics/proto"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const (
	defaultTagValue       = "N/A"
	waitingRequestSeconds = 5
)

var (
	buildVersion = defaultTagValue
	buildDate    = defaultTagValue
	buildCommit  = defaultTagValue
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
	var repository http2.AlertStorage = storage.NewInMemoryStorage()
	if err := godotenv.Load(utils.Root + "/.env-server"); err != nil {
		logger.Log.Warnf("failed load .env-server file: %v", err)
	}
	cnfg, err := config.NewServer()
	if err != nil {
		return fmt.Errorf("failed get server config: %w", err)
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

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	wg := &sync.WaitGroup{}

	if cnfg.StoreInterval > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.SaveDataToFilesystemByInterval(ctx, cnfg, repository)
		}()
	}
	if cnfg.Restore {
		if err = service.FillFromFilesystem(ctx, repository, cnfg.FilePath); err != nil {
			logger.Log.Warn(err)
		}
	}

	fmt.Println(
		"Build version: ", buildVersion, "\n",
		"Build date: ", buildDate, "\n",
		"Build commit: ", buildCommit,
	)
	logger.Log.Infof("http server is starting on port: %s...", cnfg.Host)
	serv := service.NewMetricsService(repository, cnfg)
	rt := http2.NewMetricsRouter(cnfg, serv)
	srv := http.Server{
		Addr:    cnfg.Host,
		Handler: rt.BuildRouter(),
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), waitingRequestSeconds*time.Second)
		defer timeoutCancel()
		if err = srv.Shutdown(timeoutCtx); err != nil {
			logger.Log.Errorf("filed shutdown server: %v", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Errorf("failed to start server: %w", err)
		}
	}()

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(logger.InterceptorLogger()),
	))

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		listen, err := net.Listen("tcp", cnfg.GRPCHost)
		if err != nil {
			logger.Log.Errorf("failed start grpc server: %v", err)
			return
		}

		pb.RegisterMetricsServiceServer(grpcServer, mygrpc.New(serv, cnfg))

		logger.Log.Infof("grpc server is starting on port: %s...", cnfg.GRPCHost)
		if err := grpcServer.Serve(listen); err != nil {
			logger.Log.Errorf("grpc server was crash: %v", err)
			return
		}
	}()

	wg.Wait()
	logger.Log.Info("server was gracefully shutdown.")
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
