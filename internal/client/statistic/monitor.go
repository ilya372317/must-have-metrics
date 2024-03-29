package statistic

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

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
const chunkForRequestSize = 50

// Monitor entity for collect metrics and send it to server.
type Monitor struct {
	Data         map[string]MonitorValue
	ReportTaskCh chan func()
	sync.Mutex
}

// New constructor for Monitor.
func New(poolSize uint) *Monitor {
	m := &Monitor{
		Data:         make(map[string]MonitorValue),
		ReportTaskCh: make(chan func(), poolSize),
	}
	m.startWorkerPool(poolSize)
	return m
}

func (monitor *Monitor) startWorker(workerID int) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Errorf("Worker %d recovered from panic: %v", workerID, r)
				time.Sleep(time.Second)
				monitor.startWorker(workerID)
			}
		}()
		for {
			reportTask, more := <-monitor.ReportTaskCh
			if !more {
				if len(monitor.ReportTaskCh) > 0 {
					logger.Log.Error("Some task was not completed before monitor shutdown.")
				}
				logger.Log.Infof("Worker %d is stopping because the channel is closed.", workerID)
				return
			}
			reportTask()
		}
	}()
}

func (monitor *Monitor) startWorkerPool(poolSize uint) {
	for k := 0; k < int(poolSize); k++ {
		monitor.startWorker(k)
	}
}

// MonitorValue representation of collected metric.
type MonitorValue struct {
	Name  string
	Type  string
	Value uint64
	Delta int
}

// CollectStat method for collect metrics from operating system.
func (monitor *Monitor) CollectStat(ctx context.Context, wg *sync.WaitGroup, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			monitor.collectStat()
			monitor.collectMemStat()
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

func (monitor *Monitor) collectMemStat() {
	monitor.Mutex.Lock()
	m, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Warnf("failed read virtual memory stats: %v", err)
		return
	}

	monitor.setGaugeValue("TotalMemory", m.Total)
	monitor.setGaugeValue("FreeMemory", m.Free)

	cpuPercentages, err := cpu.Percent(0, true)
	if err != nil {
		logger.Log.Warnf("failed read cpu stats: %v", err)
		return
	}
	monitor.setGaugeValue("CPUutilization1", uint64(cpuPercentages[0]))
	monitor.Mutex.Unlock()
}

func (monitor *Monitor) collectStat() {
	monitor.Mutex.Lock()
	rtm := runtime.MemStats{}
	runtime.ReadMemStats(&rtm)
	monitor.updatePollCount()
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
	monitor.setGaugeValue(randomValueName, uint64(utils.GetRandomValue(minRandomValue, maxRandomValue)))
	monitor.Mutex.Unlock()
}

func (monitor *Monitor) ReportStat(ctx context.Context, wg *sync.WaitGroup,
	agentConfig *config.AgentConfig, reportInterval time.Duration,
	reportSender sender.ReportSender) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()
	taskWg := &sync.WaitGroup{}
	for {
		select {
		case <-ticker.C:
			dataForSend := make([]MonitorValue, 0, len(monitor.Data))
			monitor.Mutex.Lock()
			for _, value := range monitor.Data {
				dataForSend = append(dataForSend, value)
			}
			monitor.Mutex.Unlock()

			dataChunks := chunkMonitorValueSlice(dataForSend, chunkForRequestSize)

			for _, chunk := range dataChunks {
				taskWg.Add(1)
				monitor.ReportTaskCh <- func() {
					defer taskWg.Done()
					monitor.reportStat(agentConfig, reportSender, chunk)
				}
			}
		case <-ctx.Done():
			taskWg.Wait()
			close(monitor.ReportTaskCh)
			wg.Done()
			return
		}
	}
}

func (monitor *Monitor) reportStat(agentConfig *config.AgentConfig,
	reportSender sender.ReportSender,
	data []MonitorValue,
) {
	requestURL := createURLForReportStat(agentConfig.Host)
	body := createBody(data)
	reportSender(agentConfig, requestURL, body)
	monitor.resetPollCount()
}

func (monitor *Monitor) setGaugeValue(name string, value uint64) {
	monitor.Data[name] = MonitorValue{
		Name:  name,
		Value: value,
		Type:  entity.TypeGauge,
	}
}

func (monitor *Monitor) updatePollCount() {
	_, ok := monitor.Data[counterName]
	if !ok {
		firstValue := 1
		monitor.Data[counterName] = MonitorValue{Name: counterName, Type: entity.TypeCounter, Delta: firstValue}
		return
	}
	oldValue := monitor.Data[counterName].Delta
	newValue := oldValue + 1
	monitor.Data[counterName] = MonitorValue{
		Name:  counterName,
		Type:  entity.TypeCounter,
		Delta: newValue,
	}
}
func (monitor *Monitor) resetPollCount() {
	nullValue := 0
	monitor.Data[counterName] = MonitorValue{
		Name:  counterName,
		Type:  entity.TypeCounter,
		Delta: nullValue,
	}
}

func createBody(data []MonitorValue) string {
	metricsList := make([]dto.Metrics, 0, len(data))
	for _, monitorValue := range data {
		m := dto.Metrics{
			ID:    monitorValue.Name,
			MType: monitorValue.Type,
		}
		if monitorValue.Type == entity.TypeCounter {
			int64Value := int64(monitorValue.Delta)
			m.Delta = &int64Value
		}
		if monitorValue.Type == entity.TypeGauge {
			float64Value := float64(monitorValue.Value)
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

func chunkMonitorValueSlice(slice []MonitorValue, chunkSize int) [][]MonitorValue {
	var chunks [][]MonitorValue
	for {
		if len(slice) == 0 {
			break
		}

		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[:chunkSize])
		slice = slice[chunkSize:]
	}
	return chunks
}
