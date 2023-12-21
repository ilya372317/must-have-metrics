package statistic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitor_collectStat(t *testing.T) {
	type want struct {
		keys []string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "success case",
			want: want{
				keys: []string{
					"Alloc", "BuckHashSys", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
					"HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
					"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "OtherSys",
					"PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "Frees",
					"NumGC", "NumForcedGC", "GCCPUFraction", "RandomValue", "PollCount",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := New()
			monitor.collectStat()

		loop:
			for {
				select {
				case value := <-monitor.DataCh:
					assert.Contains(t, tt.want.keys, value.Name)
				default:
					break loop
				}
			}
		})
	}
}
