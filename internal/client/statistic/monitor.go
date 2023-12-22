package statistic

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ilya372317/must-have-metrics/internal/client/sender"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/utils"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

const counterName = "PollCount"
const randomValueName = "RandomValue"
const minRandomValue = 1
const maxRandomValue = 50
const collectWorkerCount = 4

var counterValue = 0

type Monitor struct {
	DataCh        chan MonitorValue
	CollectTaskCh chan func()
	ReportTaskCh  chan func()
	WaitGroup     sync.WaitGroup
}

func New(rateLimit uint) *Monitor {
	m := &Monitor{
		DataCh:        make(chan MonitorValue),
		CollectTaskCh: make(chan func(), rateLimit),
		ReportTaskCh:  make(chan func(), collectWorkerCount),
	}
	m.startWorkerPool(rateLimit)
	return m
}

func (monitor *Monitor) startWorkerPool(rateLimit uint) {
	for i := 0; i < collectWorkerCount; i++ {
		monitor.WaitGroup.Add(1)
		go func() {
			defer monitor.WaitGroup.Done()
			for collectTask := range monitor.CollectTaskCh {
				collectTask()
			}
		}()
	}
	for k := 0; k < int(rateLimit); k++ {
		monitor.WaitGroup.Add(1)
		go func() {
			defer monitor.WaitGroup.Done()
			for reportTask := range monitor.ReportTaskCh {
				reportTask()
			}
		}()
	}
}

func (monitor *Monitor) Shutdown() {
	close(monitor.ReportTaskCh)
	close(monitor.CollectTaskCh)
	monitor.WaitGroup.Wait()
	close(monitor.DataCh)
}

type MonitorValue struct {
	Name  string
	Value *uint64
	Delta *int
	Type  string
}

func (monitor *Monitor) CollectStat(pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	for range ticker.C {
		monitor.CollectTaskCh <- monitor.collectStat
		monitor.CollectTaskCh <- monitor.collectMemStat
	}
}

func (monitor *Monitor) collectMemStat() {
	go func() {
		m, err := mem.VirtualMemory()
		if err != nil {
			logger.Log.Warnf("failed read virtual memory stats: %v", err)
			return
		}

		monitor.sendGaugeValue("TotalMemory", m.Total)
		monitor.sendGaugeValue("FreeMemory", m.Free)

		cpuPercentages, err := cpu.Percent(0, true)
		if err != nil {
			logger.Log.Warnf("failed read cpu stats: %v", err)
			return
		}
		monitor.sendGaugeValue("CPUutilization1", uint64(cpuPercentages[0]))
	}()
}

func (monitor *Monitor) collectStat() {
	rtm := runtime.MemStats{}
	runtime.ReadMemStats(&rtm)
	monitor.sendGaugeValue("Alloc", rtm.Alloc)
	monitor.sendGaugeValue("BuckHashSys", rtm.BuckHashSys)
	monitor.sendGaugeValue("GCSys", rtm.GCSys)
	monitor.sendGaugeValue("HeapAlloc", rtm.HeapAlloc)
	monitor.sendGaugeValue("HeapIdle", rtm.HeapIdle)
	monitor.sendGaugeValue("HeapInuse", rtm.HeapInuse)
	monitor.sendGaugeValue("HeapObjects", rtm.HeapObjects)
	monitor.sendGaugeValue("HeapReleased", rtm.HeapReleased)
	monitor.sendGaugeValue("HeapSys", rtm.HeapSys)
	monitor.sendGaugeValue("LastGC", rtm.LastGC)
	monitor.sendGaugeValue("Lookups", rtm.Lookups)
	monitor.sendGaugeValue("MCacheInuse", rtm.MCacheInuse)
	monitor.sendGaugeValue("MCacheSys", rtm.MCacheSys)
	monitor.sendGaugeValue("MSpanInuse", rtm.MSpanInuse)
	monitor.sendGaugeValue("MSpanSys", rtm.MSpanSys)
	monitor.sendGaugeValue("Mallocs", rtm.Mallocs)
	monitor.sendGaugeValue("NextGC", rtm.NextGC)
	monitor.sendGaugeValue("OtherSys", rtm.OtherSys)
	monitor.sendGaugeValue("PauseTotalNs", rtm.PauseTotalNs)
	monitor.sendGaugeValue("StackInuse", rtm.StackInuse)
	monitor.sendGaugeValue("StackSys", rtm.StackSys)
	monitor.sendGaugeValue("Sys", rtm.Sys)
	monitor.sendGaugeValue("TotalAlloc", rtm.TotalAlloc)
	monitor.sendGaugeValue("Frees", rtm.Frees)
	monitor.sendGaugeValue("NumGC", uint64(rtm.NumGC))
	monitor.sendGaugeValue("NumForcedGC", uint64(rtm.NumForcedGC))
	monitor.sendGaugeValue("GCCPUFraction", uint64(rtm.GCCPUFraction))
	monitor.sendGaugeValue(randomValueName, uint64(utils.GetRandomValue(minRandomValue, maxRandomValue)))
	monitor.updatePollCount()
}

func (monitor *Monitor) ReportStat(agentConfig *config.AgentConfig, reportInterval time.Duration,
	reportSender sender.ReportSender) {
	ticker := time.NewTicker(reportInterval)
	for range ticker.C {
		for i := 0; i < int(agentConfig.RateLimit); i++ {
			monitor.ReportTaskCh <- func() {
				monitor.reportStat(agentConfig, reportSender)
			}
		}
	}
}

func (monitor *Monitor) reportStat(agentConfig *config.AgentConfig, reportSender sender.ReportSender) {
	data := make([]MonitorValue, 0, len(monitor.DataCh))
loop:
	for {
		select {
		case value := <-monitor.DataCh:
			data = append(data, value)
		default:
			break loop
		}
	}

	if len(data) == 0 {
		return
	}
	requestURL := createURLForReportStat(agentConfig.Host)
	body := createBody(data)
	reportSender(agentConfig, requestURL, body)
	monitor.resetPollCount()
}

func (monitor *Monitor) sendGaugeValue(name string, value uint64) {
	monitor.DataCh <- MonitorValue{
		Name:  name,
		Value: &value,
		Type:  entity.TypeGauge,
	}
}

func (monitor *Monitor) updatePollCount() {
	value := int(atomic.AddInt64((*int64)(unsafe.Pointer(&counterValue)), 1))
	monitor.DataCh <- MonitorValue{
		Name:  counterName,
		Delta: &value,
		Type:  entity.TypeCounter,
	}
}

func (monitor *Monitor) resetPollCount() {
	atomic.StoreInt64((*int64)(unsafe.Pointer(&counterValue)), 0)
}

func createBody(data []MonitorValue) string {
	metricsList := make([]dto.Metrics, 0, len(data))
	for _, monitorValue := range data {
		m := dto.Metrics{
			ID:    monitorValue.Name,
			MType: monitorValue.Type,
		}
		if monitorValue.Type == entity.TypeCounter {
			int64Value := int64(*monitorValue.Delta)
			m.Delta = &int64Value
		}
		if monitorValue.Type == entity.TypeGauge {
			float64Value := float64(*monitorValue.Value)
			m.Value = &float64Value
		}
		metricsList = append(metricsList, m)
	}

	body, _ := json.Marshal(&metricsList)
	return string(body)
}

func createURLForReportStat(host string) string {
	return fmt.Sprintf("http://" + host + "/updates")
}
