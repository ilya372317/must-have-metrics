package storage

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func TestDatabaseStorage_Save(t *testing.T) {
	type args struct {
		name  string
		alert testAlert
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success case",
			args: args{
				name: "alert",
				alert: testAlert{
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1.1,
					IntValue:   0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open(
				"pgx",
				"host=localhost user=ilya password=Ilya372317 dbname=metrics sslmode=disable",
			)
			require.NoError(t, err)
			d := &DatabaseStorage{
				db: db,
			}
			d.Save(tt.args.name, newAlertFromTestAlert(tt.args.alert))
		})
	}
}
