package statistic

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/utils"
)

const counterName = "PollCount"
const randomValueName = "RandomValue"
const minRandomValue = 1
const maxRandomValue = 50

type Monitor struct {
	Data map[string]MonitorValue
	sync.Mutex
}

func New() Monitor {
	return Monitor{Data: make(map[string]MonitorValue)}
}

type MonitorValue struct {
	Value *uint64
	Delta *int
	Type  string
}

func (monitor *Monitor) CollectStat(pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	rtm := runtime.MemStats{}
	for range ticker.C {
		runtime.ReadMemStats(&rtm)
		monitor.Lock()
		monitor.collectStat(&rtm)
		monitor.Unlock()
	}
}

func (monitor *Monitor) collectStat(rtm *runtime.MemStats) {
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
	monitor.setGaugeValue("NumGC", uint64(rtm.NumGC))
	monitor.setGaugeValue("NumForcedGC", uint64(rtm.NumForcedGC))
	monitor.setGaugeValue("GCCPUFraction", uint64(rtm.GCCPUFraction))
	monitor.setCounterValue(randomValueName, utils.GetRandomValue(minRandomValue, maxRandomValue))
	monitor.updatePollCount()
}

func (monitor *Monitor) ReportStat(host string, reportInterval time.Duration, reportSender sender.ReportSender) {
	ticker := time.NewTicker(reportInterval)
	for range ticker.C {
		monitor.Lock()
		monitor.reportStat(host, reportSender)
		monitor.Unlock()
	}
}

func (monitor *Monitor) reportStat(host string, reportSender sender.ReportSender) {
	for statName, data := range monitor.Data {
		requestURL := createURLForReportStat(host)
		body := createBody(statName, data)
		reportSender(requestURL, body)
	}
	monitor.resetPollCount()
}

func (monitor *Monitor) setGaugeValue(name string, value uint64) {
	monitor.Data[name] = MonitorValue{
		Type:  entity.TypeGauge,
		Value: &value,
	}
}

func (monitor *Monitor) setCounterValue(name string, value int) {
	monitor.Data[name] = MonitorValue{
		Type:  entity.TypeCounter,
		Delta: &value,
	}
}

func (monitor *Monitor) updatePollCount() {
	_, ok := monitor.Data[counterName]
	if !ok {
		firstValue := 1
		monitor.Data[counterName] = MonitorValue{Type: entity.TypeCounter, Delta: &firstValue}
		return
	}
	oldValue := monitor.Data[counterName].Delta
	newValue := *oldValue + 1
	monitor.Data[counterName] = MonitorValue{
		Type:  entity.TypeCounter,
		Delta: &newValue,
	}
}
func (monitor *Monitor) resetPollCount() {
	nullValue := 0
	monitor.Data[counterName] = MonitorValue{
		Type:  entity.TypeCounter,
		Delta: &nullValue,
	}
}

func createURLForReportStat(host string) string {
	return fmt.Sprintf("http://" + host + "/update")
}

func createBody(name string, monitorValue MonitorValue) string {
	metrics := dto.Metrics{
		ID:    name,
		MType: monitorValue.Type,
	}
	if monitorValue.Type == entity.TypeCounter {
		int64Value := int64(*monitorValue.Delta)
		metrics.Delta = &int64Value
	}
	if monitorValue.Type == entity.TypeGauge {
		float64Value := float64(*monitorValue.Value)
		metrics.Value = &float64Value
	}

	body, _ := json.Marshal(&metrics)
	return string(body)
}
