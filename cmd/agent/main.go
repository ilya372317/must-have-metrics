package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type Monitor struct {
	sync.Mutex
	Data map[string]interface{}
}

func main() {
	monitor := Monitor{Data: make(map[string]interface{})}
	pollInterval := time.Second * 2
	reportInterval := time.Second * 10

	go CollectStat(&monitor, pollInterval)
	go ReportStat(&monitor, reportInterval)
	select {}
}

func CollectStat(monitor *Monitor, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	rtm := runtime.MemStats{}
	for range ticker.C {
		runtime.ReadMemStats(&rtm)
		monitor.Lock()
		monitor.Data["Alloc"] = rtm.Alloc
		monitor.Data["BuckHashSys"] = rtm.BuckHashSys
		monitor.Data["GCSys"] = rtm.GCSys
		monitor.Data["HeapAlloc"] = rtm.HeapAlloc
		monitor.Data["HeapIdle"] = rtm.HeapIdle
		monitor.Data["HeapInuse"] = rtm.HeapInuse
		monitor.Data["HeapObjects"] = rtm.HeapObjects
		monitor.Data["HeapReleased"] = rtm.HeapReleased
		monitor.Data["HeapSys"] = rtm.HeapSys
		monitor.Data["LastGC"] = rtm.LastGC
		monitor.Data["Lookups"] = rtm.Lookups
		monitor.Data["MCacheInuse"] = rtm.MCacheInuse
		monitor.Data["MCacheSys"] = rtm.MCacheSys
		monitor.Data["MSpanInuse"] = rtm.MSpanInuse
		monitor.Data["MSpanSys"] = rtm.MSpanSys
		monitor.Data["Mallocs"] = rtm.Mallocs
		monitor.Data["NextGC"] = rtm.NextGC
		monitor.Data["OtherSys"] = rtm.OtherSys
		monitor.Data["PauseTotalNs"] = rtm.PauseTotalNs
		monitor.Data["StackInuse"] = rtm.StackInuse
		monitor.Data["StackSys"] = rtm.StackSys
		monitor.Data["Sys"] = rtm.Sys
		monitor.Data["TotalAlloc"] = rtm.TotalAlloc
		monitor.Data["Frees"] = rtm.Frees
		monitor.Data["NumGC"] = rtm.NumGC
		monitor.Data["NumForcedGC"] = rtm.NumForcedGC
		monitor.Data["GCCPUFraction"] = rtm.GCCPUFraction
		monitor.Unlock()
	}
}

func ReportStat(monitor *Monitor, reportInterval time.Duration) {
	ticker := time.NewTicker(reportInterval)
	for range ticker.C {
		monitor.Lock()
		for statName, value := range monitor.Data {
			requestURL := fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", "gauge", statName, value)
			if _, err := http.Post(requestURL, "text/plain", nil); err != nil {
				log.Printf("failed to save data on server: %v\n", err)
			}
		}
		monitor.Unlock()
	}
}
