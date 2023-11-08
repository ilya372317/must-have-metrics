package main

import (
	"flag"
	"log"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/client/statistic"
	"github.com/ilya372317/must-have-metrics/internal/config"
)

var (
	host           *string
	pollInterval   *int
	reportInterval *int
)

func init() {
	cnfg := new(config.AgentConfig)
	if err := cnfg.Init(); err != nil {
		log.Fatalln(err.Error())
	}
	host = flag.String("a", "localhost:8080", "server address")
	pollInterval = flag.Int("p", 2, "frequency of metrics collection")
	reportInterval = flag.Int("r", 10, "frequency of send metrics on server")

	if cnfg.Host != "" {
		host = &cnfg.Host
	}
	if cnfg.PollInterval != 0 {
		pollInterval = &cnfg.PollInterval
	}
	if cnfg.ReportInterval != 0 {
		reportInterval = &cnfg.ReportInterval
	}
}

func main() {
	flag.Parse()
	monitor := statistic.New()
	go monitor.CollectStat(time.Duration(*pollInterval) * time.Second)
	go monitor.ReportStat(*host, time.Duration(*reportInterval)*time.Second, sender.SendReport)
	select {}
}
