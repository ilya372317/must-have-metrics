package main

import (
	"fmt"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/client/statistic"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(fmt.Errorf("failed init logger for agent: %w", err))
	}
	cnfg, err := config.NewAgent()
	if err != nil {
		logger.Log.Panicf("failed get config: %v", err)
	}
	monitor := statistic.New()
	go monitor.CollectStat(time.Duration(cnfg.PollInterval) * time.Second)
	go monitor.ReportStat(
		cnfg.Host,
		time.Duration(cnfg.ReportInterval)*time.Second,
		sender.SendReport,
	)
	select {}
}
