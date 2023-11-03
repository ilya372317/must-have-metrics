package main

import (
	"flag"
	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/client/statistic"
	"time"
)

var (
	serverHost     *string
	pollInterval   *int
	reportInterval *int
)

func init() {
	serverHost = flag.String("a", "localhost:8080", "server address")
	pollInterval = flag.Int("p", 2, "frequency of metrics collection")
	reportInterval = flag.Int("r", 10, "frequency of send metrics on server")
}

func main() {
	flag.Parse()
	monitor := statistic.Monitor{Data: make(map[string]statistic.MonitorValue)}

	go monitor.CollectStat(time.Duration(*pollInterval) * time.Second)
	go monitor.ReportStat(*serverHost, time.Duration(*reportInterval)*time.Second, sender.SendReport)
	select {}
}
