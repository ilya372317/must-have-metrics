package statistic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"runtime"
	"testing"
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
			rtm := runtime.MemStats{}
			runtime.ReadMemStats(&rtm)
			monitor.collectStat(&rtm)

			for _, statName := range tt.want.keys {
				_, ok := monitor.Data[statName]
				assert.True(t, ok)
			}

			pollCount, pollCountExist := monitor.Data["PollCount"]
			require.True(t, pollCountExist)
			assert.Equal(t, 1, pollCount.Value)
			randomValue, randomValueExist := monitor.Data["RandomValue"]
			require.True(t, randomValueExist)
			if randomValue.Value.(int) <= 0 {
				t.Errorf("invalid random value")
			}
		})
	}
}
