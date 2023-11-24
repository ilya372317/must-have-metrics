package main

import (
	"strconv"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/client/statistic"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var (
	agentLogger = logger.Get()
)

func main() {
	if err := config.InitAgentConfig(); err != nil {
		agentLogger.Panicf("failed init agent config: %v", err)
	}

	cnfg := config.GetAgentConfig()
	pollInterval, err := strconv.Atoi(cnfg.GetValue("poll_interval"))
	if err != nil {
		agentLogger.Panicf("failed parse poll interval: %v", err)
	}
	reportInterval, err := strconv.Atoi(cnfg.GetValue("report_interval"))
	if err != nil {
		agentLogger.Panicf("failed parse report interval: %v", err)
	}

	monitor := statistic.New()
	go monitor.CollectStat(time.Duration(pollInterval) * time.Second)
	go monitor.ReportStat(
		cnfg.GetValue("host"),
		time.Duration(reportInterval)*time.Second,
		sender.SendReport,
	)
	select {}
}
