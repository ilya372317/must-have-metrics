package service

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var serverConfig = &config.ServerConfig{
	Host:          "localhost:8080",
	FilePath:      "/tmp/metrics.json",
	Restore:       true,
	StoreInterval: 300,
}

type testAlert struct {
	Type       string
	Name       string
	FloatValue float64
	IntValue   int64
}

func Test_addAlert(t *testing.T) {
	type args struct {
		repo *storage.InMemoryStorage
		dto  dto.Metrics
	}
	tests := []struct {
		fields  map[string]testAlert
		name    string
		args    args
		want    testAlert
		wantErr bool
	}{
		{
			name: "success counter empty storage case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.Metrics{
					MType: "counter",
					ID:    "alert",
					Delta: intPointer(10),
				},
			},
			fields:  map[string]testAlert{},
			wantErr: false,
			want: testAlert{
				Type:     "counter",
				Name:     "alert",
				IntValue: int64(10),
			},
		},
		{
			name: "success counter not empty storage case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.Metrics{
					MType: "counter",
					ID:    "alert",
					Delta: intPointer(10),
				},
			},
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(20),
				},
			},
			wantErr: false,
			want: testAlert{
				Type:     "counter",
				Name:     "alert",
				IntValue: int64(30),
			},
		},
		{
			name: "success gauge empty storage case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.Metrics{
					MType: "gauge",
					ID:    "alert",
					Value: floatPointer(1.1),
				},
			},
			fields:  map[string]testAlert{},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1.1,
			},
		},
		{
			name: "success gauge not empty storage case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.Metrics{
					MType: "gauge",
					ID:    "alert",
					Value: floatPointer(1.2),
				},
			},
			fields: map[string]testAlert{
				"alert": {
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1.1,
				},
			},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1.2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				err := tt.args.repo.Save(context.Background(), name, newAlertFromTestAlert(alert))
				require.NoError(t, err)
			}

			_, err := AddAlert(context.Background(), tt.args.repo, tt.args.dto, serverConfig)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			addedAlert, addAlertErr := tt.args.repo.Get(context.Background(), tt.args.dto.ID)
			require.NoError(t, addAlertErr)
			assert.Equal(t, addedAlert, newAlertFromTestAlert(tt.want))
		})
	}
}

func Test_updateCounterAlert(t *testing.T) {
	type args struct {
		repo *storage.InMemoryStorage
		dto  dto.Metrics
	}
	tests := []struct {
		fields  map[string]testAlert
		name    string
		args    args
		want    testAlert
		wantErr bool
	}{
		{
			name: "success case with empty storage",
			args: args{
				dto: dto.Metrics{
					MType: "counter",
					ID:    "alert",
					Delta: intPointer(10),
				},
				repo: storage.NewInMemoryStorage(),
			},
			fields:  map[string]testAlert{},
			wantErr: false,
			want: testAlert{
				Type:     "counter",
				Name:     "alert",
				IntValue: int64(10),
			},
		},
		{
			name: "success case with value in storage",
			args: args{
				dto: dto.Metrics{
					MType: "counter",
					ID:    "alert",
					Delta: intPointer(10),
				},
				repo: storage.NewInMemoryStorage(),
			},
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(10),
				},
			},
			wantErr: false,
			want: testAlert{
				Type:     "counter",
				Name:     "alert",
				IntValue: int64(20),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				err := tt.args.repo.Save(context.Background(), name, newAlertFromTestAlert(alert))
				require.NoError(t, err)
			}

			_, err := updateCounterAlert(context.Background(), tt.args.dto, tt.args.repo)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			updatedAlert, getAlertErr := tt.args.repo.Get(context.Background(), tt.args.dto.ID)
			require.NoError(t, getAlertErr)
			assert.Equal(t, updatedAlert, newAlertFromTestAlert(tt.want))
		})
	}
}

func Test_updateGaugeAlert(t *testing.T) {
	type args struct {
		repository *storage.InMemoryStorage
		dto        dto.Metrics
	}
	tests := []struct {
		name    string
		args    args
		want    testAlert
		wantErr bool
	}{
		{
			name: "positive test",
			args: args{
				dto: dto.Metrics{
					MType: "gauge",
					ID:    "alert",
					Value: floatPointer(1.12),
				},
				repository: storage.NewInMemoryStorage(),
			},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1.12,
			},
		},
		{
			name: "parse integer value",
			args: args{
				dto: dto.Metrics{
					MType: "gauge",
					ID:    "alert",
					Value: floatPointer(1),
				},
				repository: storage.NewInMemoryStorage(),
			},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := updateGaugeAlert(context.Background(), tt.args.dto, tt.args.repository)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			expectedAlert, getAlertError := tt.args.repository.Get(context.Background(), tt.args.dto.ID)
			require.NoError(t, getAlertError)
			assert.Equal(t, newAlertFromTestAlert(tt.want), expectedAlert)
		})
	}
}

func newAlertFromTestAlert(testAlert testAlert) entity.Alert {
	wantAlert := entity.Alert{
		Type: testAlert.Type,
		Name: testAlert.Name,
	}
	if testAlert.FloatValue != 0 {
		floatValue := testAlert.FloatValue
		wantAlert.FloatValue = &floatValue
	}
	if testAlert.IntValue != 0 {
		intValue := testAlert.IntValue
		wantAlert.IntValue = &intValue
	}

	return wantAlert
}

func Test_FillAndSaveFromFile(t *testing.T) {
	tests := []struct {
		name           string
		filepath       string
		items          []testAlert
		wantFillErr    bool
		wantRestoreErr bool
	}{
		{
			name:           "success empty storage case",
			items:          nil,
			filepath:       "test-metrics.json",
			wantFillErr:    false,
			wantRestoreErr: false,
		},
		{
			name: "success complex case",
			items: []testAlert{
				{
					FloatValue: 1.1,
					Type:       "gauge",
					Name:       "alert1",
				},
				{
					IntValue: int64(1),
					Type:     "counter",
					Name:     "alert2",
				},
				{
					FloatValue: 1.234567,
					Type:       "gauge",
					Name:       "alert3",
				},
				{
					IntValue: int64(123456),
					Type:     "counter",
					Name:     "alert4",
				},
			},
			filepath:       "test-metrics.json",
			wantFillErr:    false,
			wantRestoreErr: false,
		},
		{
			name:           "negative empty file path case",
			items:          nil,
			filepath:       "",
			wantFillErr:    true,
			wantRestoreErr: false,
		},
	}

	for _, tt := range tests {
		memoryStorage := storage.NewInMemoryStorage()
		t.Run(tt.name, func(t *testing.T) {
			for _, alert := range tt.items {
				err := memoryStorage.Save(context.Background(), alert.Name, newAlertFromTestAlert(alert))
				require.NoError(t, err)
			}

			errStore := StoreToFilesystem(context.Background(), memoryStorage, tt.filepath)
			if tt.wantFillErr {
				assert.Error(t, errStore)
				return
			} else {
				require.NoError(t, errStore)
			}
			expect, _ := memoryStorage.All(context.Background())

			memoryStorage.Reset()

			errFill := FillFromFilesystem(context.Background(), memoryStorage, tt.filepath)
			if tt.wantRestoreErr {
				assert.Error(t, errFill)
				return
			} else {
				require.NoError(t, errFill)
			}

			got, _ := memoryStorage.All(context.Background())
			sort.SliceStable(expect, func(i, j int) bool {
				return expect[i].Name > expect[j].Name
			})
			sort.SliceStable(got, func(i, j int) bool {
				return got[i].Name > got[j].Name
			})

			assert.Equal(t, expect, got)

			_ = os.Remove(tt.filepath)
			memoryStorage.Reset()
		})
	}
}

func intPointer(value int64) *int64 {
	return &value
}

func floatPointer(value float64) *float64 {
	return &value
}
