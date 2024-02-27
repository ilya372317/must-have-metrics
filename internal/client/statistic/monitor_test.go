package statistic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			monitor := New(10)
			monitor.collectStat()

			for _, statName := range tt.want.keys {
				_, ok := monitor.Data[statName]
				assert.True(t, ok)
			}

			pollCount, pollCountExist := monitor.Data["PollCount"]
			require.True(t, pollCountExist)
			assert.Equal(t, 1, *pollCount.Delta)
			randomValue, randomValueExist := monitor.Data["RandomValue"]
			require.True(t, randomValueExist)
			if *randomValue.Value <= 0 {
				t.Errorf("invalid random value")
			}
		})
	}
}

func BenchmarkMonitor_collectStat1(b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()
	m := Monitor{
		Data:         make(map[string]MonitorValue),
		ReportTaskCh: make(chan func(), 10),
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		m.collectStat()
	}
}
