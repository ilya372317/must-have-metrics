package statistic

import (
	"fmt"
	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/constant"
	"github.com/ilya372317/must-have-metrics/internal/utils"
	"runtime"
	"sync"
	"time"
)

const counterName = "PollCount"
const randomValueName = "RandomValue"
const minRandomValue = 1
const maxRandomValue = 50
const counterStep = 1

type Monitor struct {
	sync.Mutex
	Data map[string]MonitorValue
}

type MonitorValue struct {
	Type  string
	Value interface{}
}

func (monitor *Monitor) CollectStat(pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	rtm := runtime.MemStats{}
	for range ticker.C {
		runtime.ReadMemStats(&rtm)
		monitor.Lock()
		monitor.setGaugeValue("Alloc", rtm.Alloc)
		monitor.setGaugeValue("BuckHashSys", rtm.BuckHashSys)
		monitor.setGaugeValue("GCSys", rtm.GCSys)
		monitor.setGaugeValue("HeapAlloc", rtm.HeapAlloc)
		monitor.setGaugeValue("HeapIdle", rtm.HeapIdle)
		monitor.setGaugeValue("HeapInuse", rtm.HeapInuse)
		monitor.setGaugeValue("HeapObjects", rtm.HeapObjects)
		monitor.setGaugeValue("HeapReleased", rtm.HeapReleased)
		monitor.setGaugeValue("HeapSys", rtm.HeapSys)
		monitor.setGaugeValue("LastGC", rtm.LastGC)
		monitor.setGaugeValue("Lookups", rtm.Lookups)
		monitor.setGaugeValue("MCacheInuse", rtm.MCacheInuse)
		monitor.setGaugeValue("MCacheSys", rtm.MCacheSys)
		monitor.setGaugeValue("MSpanInuse", rtm.MSpanInuse)
		monitor.setGaugeValue("MSpanSys", rtm.MSpanSys)
		monitor.setGaugeValue("Mallocs", rtm.Mallocs)
		monitor.setGaugeValue("NextGC", rtm.NextGC)
		monitor.setGaugeValue("OtherSys", rtm.OtherSys)
		monitor.setGaugeValue("PauseTotalNs", rtm.PauseTotalNs)
		monitor.setGaugeValue("StackInuse", rtm.StackInuse)
		monitor.setGaugeValue("StackSys", rtm.StackSys)
		monitor.setGaugeValue("Sys", rtm.Sys)
		monitor.setGaugeValue("TotalAlloc", rtm.TotalAlloc)
		monitor.setGaugeValue("Frees", rtm.Frees)
		monitor.setGaugeValue("NumGC", rtm.NumGC)
		monitor.setGaugeValue("NumForcedGC", rtm.NumForcedGC)
		monitor.setGaugeValue("GCCPUFraction", rtm.GCCPUFraction)
		monitor.setCounterValue(randomValueName, utils.GetRandomValue(minRandomValue, maxRandomValue))
		monitor.setCounterValue(counterName, counterStep)
		monitor.Unlock()
	}
}

func (monitor *Monitor) ReportStat(reportInterval time.Duration, reportSender sender.ReportSender) {
	ticker := time.NewTicker(reportInterval)
	for range ticker.C {
		monitor.Lock()
		for statName, data := range monitor.Data {
			requestURL := createURLForReportStat(data.Type, statName, data.Value)
			reportSender(requestURL)
		}
		monitor.Unlock()
	}
}

func (monitor *Monitor) setGaugeValue(name string, value interface{}) {
	monitor.Data[name] = MonitorValue{
		Type:  constant.TypeGauge,
		Value: value,
	}
}

func (monitor *Monitor) setCounterValue(name string, value interface{}) {
	monitor.Data[name] = MonitorValue{
		Type:  constant.TypeCounter,
		Value: value,
	}
}

func createURLForReportStat(typ, name string, value interface{}) string {
	return fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", typ, name, value)
}
