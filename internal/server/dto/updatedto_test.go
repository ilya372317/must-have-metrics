package dto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateAlertDTO_Validate(t *testing.T) {
	type fields struct {
		Type string
		Name string
		Data string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "success gauge case",
			fields: fields{
				Type: "gauge",
				Name: "alert",
				Data: "1.1",
			},
			want: true,
		},
		{
			name: "success counter case",
			fields: fields{
				Type: "counter",
				Name: "alert",
				Data: "1",
			},
			want: true,
		},
		{
			name: "type is invalid case",
			fields: fields{
				Type: "invalidtype",
				Name: "alert",
				Data: "1.1",
			},
			want: false,
		},
		{
			name: "invalid name case",
			fields: fields{
				Type: "gauge",
				Name: "",
				Data: "1.1",
			},
			want: false,
		},
		{
			name: "invalid value case",
			fields: fields{
				Type: "gauge",
				Name: "alert",
				Data: "invaliddata",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := &UpdateAlertDTO{
				Type: tt.fields.Type,
				Name: tt.fields.Name,
				Data: tt.fields.Data,
			}
			got, err := dto.Validate()
			if got != tt.want {
				t.Errorf("Validate() got = %v, want %v", got, tt.want)
			}
			if tt.want {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
