// Application for retrieving metrics from the operating system and sending them to a server
package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/client/statistic"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/utils"
	"github.com/joho/godotenv"
)

const defaultTagValue = "N/A"

var (
	buildVersion = defaultTagValue
	buildDate    = defaultTagValue
	buildCommit  = defaultTagValue
)

func main() {
	if err := logger.Init(); err != nil {
		panic(fmt.Errorf("failed init logger for agent: %w", err))
	}
	if err := godotenv.Load(utils.Root + "/.env-agent"); err != nil {
		logger.Log.Warnf("failed load .env-agent file: %v", err)
	}
	cnfg, err := config.NewAgent()
	if err != nil {
		logger.Log.Panicf("failed get config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	monitor := statistic.New(cnfg.RateLimit)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go monitor.CollectStat(ctx, wg, time.Duration(cnfg.PollInterval)*time.Second)
	wg.Add(1)
	go monitor.ReportStat(
		ctx,
		wg,
		cnfg,
		time.Duration(cnfg.ReportInterval)*time.Second,
		sender.SendReport,
	)
	fmt.Println(
		"Build version: ", buildVersion, "\n",
		"Build date: ", buildDate, "\n",
		"Build commit: ", buildCommit,
	)
	wg.Wait()
}
