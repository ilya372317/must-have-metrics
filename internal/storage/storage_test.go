package storage

import (
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testAlert struct {
	Type       string
	Name       string
	FloatValue float64
	IntValue   int64
}

func TestInMemoryStorage_GetAlert(t *testing.T) {
	type fields struct {
		Records map[string]testAlert
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    testAlert
		wantErr bool
	}{
		{
			name: "positive test",
			fields: fields{Records: map[string]testAlert{
				"alert": {
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1,
				},
			}},
			args: args{name: "alert"},
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1,
			},
			wantErr: false,
		},
		{
			name:    "negative case",
			fields:  fields{Records: map[string]testAlert{}},
			args:    args{name: "test"},
			want:    testAlert{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &InMemoryStorage{
				Records: makeRecordsForStorage(tt.fields.Records),
			}
			got, err := storage.Get(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, newAlertFromTestAlert(tt.want)) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryStorage_HasAlert(t *testing.T) {
	type fields struct {
		Records map[string]testAlert
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "positive test",
			fields: fields{Records: map[string]testAlert{
				"test": {
					Type:       "gauge",
					Name:       "test",
					FloatValue: 10,
				},
			}},
			args: args{name: "test"},
			want: true,
		},
		{
			name:   "negative test",
			fields: fields{Records: map[string]testAlert{}},
			args:   args{name: "test"},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &InMemoryStorage{
				Records: makeRecordsForStorage(tt.fields.Records),
			}
			if got := storage.Has(tt.args.name); got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryStorage_SaveAlert(t *testing.T) {
	type args struct {
		name  string
		alert testAlert
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "simple success test case",
			args: args{
				name: "testValue",
				alert: testAlert{
					Type:       "gauge",
					Name:       "test",
					FloatValue: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()
			storage.Save(tt.args.name, newAlertFromTestAlert(tt.args.alert))
			value, hasRecord := storage.Records[tt.args.name]
			expectedAlert := newAlertFromTestAlert(tt.args.alert)
			assert.Equal(t, expectedAlert.GetValue(), value.GetValue())
			assert.True(t, hasRecord)
		})
	}
}

func TestInMemoryStorage_UpdateAlert(t *testing.T) {
	type fields struct {
		Records map[string]testAlert
	}
	type args struct {
		name     string
		newValue testAlert
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    testAlert
	}{
		{
			name: "positive case",
			fields: fields{Records: map[string]testAlert{
				"alert": {
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1,
				},
			}},
			args: args{
				name: "alert",
				newValue: testAlert{
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 2,
				},
			},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 2,
			},
		},
		{
			name: "negative case",
			fields: fields{Records: map[string]testAlert{
				"alert_test": {
					Type:       "gauge",
					Name:       "alert_test",
					FloatValue: 1,
				},
			}},
			args: args{
				name: "alert",
				newValue: testAlert{
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 10,
				},
			},
			wantErr: true,
			want:    testAlert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()
			storage.Records = makeRecordsForStorage(tt.fields.Records)
			err := storage.Update(tt.args.name, newAlertFromTestAlert(tt.args.newValue))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, storage.Records[tt.args.name], newAlertFromTestAlert(tt.want))
		})
	}
}

func TestInMemoryStorage_FillAndSaveFromFile(t *testing.T) {
	tests := []struct {
		name           string
		items          []testAlert
		filepath       string
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
		storage := NewInMemoryStorage()
		t.Run(tt.name, func(t *testing.T) {
			for _, alert := range tt.items {
				storage.Save(alert.Name, newAlertFromTestAlert(alert))
			}

			errStore := storage.StoreToFilesystem(tt.filepath)
			if tt.wantFillErr {
				assert.Error(t, errStore)
				return
			} else {
				require.NoError(t, errStore)
			}
			expect := storage.All()

			storage.Reset()

			errFill := storage.FillFromFilesystem(tt.filepath)
			if tt.wantRestoreErr {
				assert.Error(t, errFill)
				return
			} else {
				require.NoError(t, errFill)
			}

			got := storage.All()
			sort.SliceStable(expect, func(i, j int) bool {
				return expect[i].Name > expect[j].Name
			})
			sort.SliceStable(got, func(i, j int) bool {
				return got[i].Name > got[j].Name
			})

			assert.Equal(t, expect, got)

			_ = os.Remove(tt.filepath)
			storage.Reset()
		})
	}
}

func makeRecordsForStorage(testRecords map[string]testAlert) map[string]entity.Alert {
	records := make(map[string]entity.Alert)
	for name, tAlert := range testRecords {
		newAlert := entity.Alert{
			Type: tAlert.Type,
			Name: tAlert.Name,
		}
		if tAlert.FloatValue != 0 {
			floatValue := tAlert.FloatValue
			newAlert.FloatValue = &floatValue
		}
		if tAlert.IntValue != 0 {
			intValue := tAlert.IntValue
			newAlert.IntValue = &intValue
		}
		records[name] = newAlert
	}

	return records
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
