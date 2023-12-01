package storage

import (
	"reflect"
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
