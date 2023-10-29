package main

import (
	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/client/statistic"
	"time"
)

func main() {
	monitor := statistic.Monitor{Data: make(map[string]statistic.MonitorValue)}
	pollInterval := time.Second * 2
	reportInterval := time.Second * 10

	go monitor.CollectStat(pollInterval)
	go monitor.ReportStat(reportInterval, sender.SendReport)
	select {}
}
