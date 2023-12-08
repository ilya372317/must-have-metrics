package storage

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseStorage_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	d := &DatabaseStorage{db: db}

	floatValue := 1.1
	alert := entity.Alert{
		Type:       "gauge",
		Name:       "test_alert",
		FloatValue: &floatValue,
		IntValue:   nil,
	}

	mock.ExpectExec(
		regexp.QuoteMeta(`INSERT INTO metrics ("id", "type", "int_value", "float_value") VALUES ($1,$2,$3,$4)`),
	).
		WithArgs(alert.Name, alert.Type, alert.IntValue, alert.FloatValue).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = d.Save(context.Background(), alert.Name, alert)
	require.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDatabaseStorage_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()
	alert := entity.Alert{
		Type:       "gauge",
		Name:       "alert",
		FloatValue: nil,
		IntValue:   nil,
	}
	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE metrics SET "type" = $1, "float_value" = $2, "int_value" = $3 WHERE id = $4`),
	).
		WithArgs(alert.Type, alert.FloatValue, alert.IntValue, alert.Name).
		WillReturnResult(sqlmock.NewResult(1, 1))

	d := DatabaseStorage{db: db}
	err = d.Update(context.Background(), "alert", alert)
	require.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDatabaseStorage_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	columns := []string{"id", "type", "float_value", "int_value"}

	expectedAlert := entity.Alert{
		Name:       "test_alert",
		Type:       "gauge",
		FloatValue: nil,
		IntValue:   new(int64),
	}
	*expectedAlert.IntValue = 10
	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT "id", "type", "float_value", "int_value" FROM metrics WHERE id = $1`),
	).
		WithArgs(expectedAlert.Name).
		WillReturnRows(
			sqlmock.NewRows(columns).
				AddRow(expectedAlert.Name, expectedAlert.Type, expectedAlert.FloatValue, expectedAlert.IntValue),
		).
		RowsWillBeClosed()

	d := DatabaseStorage{db: db}
	resultAlert, err := d.Get(context.Background(), expectedAlert.Name)

	require.NoError(t, err)

	assert.Equal(t, expectedAlert, resultAlert)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDatabaseStorage_Has(t *testing.T) {
	type want struct {
		expect  bool
		wantErr bool
	}
	tests := []struct {
		name     string
		want     want
		argument string
		fields   []string
	}{
		{
			name: "success simple case",
			want: want{
				expect:  true,
				wantErr: false,
			},
			argument: "alert",
			fields:   []string{"alert", "alert1"},
		},
		{
			name: "negative case",
			want: want{
				expect:  false,
				wantErr: false,
			},
			argument: "alert",
			fields:   []string{},
		},
		{
			name: "negative case with existing records",
			want: want{
				expect:  false,
				wantErr: false,
			},
			argument: "alert",
			fields:   []string{"alert1", "alert2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			rows := sqlmock.NewRows([]string{"id"})

			if contains(tt.fields, tt.argument) {
				for _, id := range tt.fields {
					rows.AddRow(id)
				}
			}

			mock.ExpectQuery(
				regexp.QuoteMeta(`SELECT "id" FROM metrics WHERE id = $1`),
			).
				WithArgs(tt.argument).
				WillReturnRows(rows).
				RowsWillBeClosed()

			d := DatabaseStorage{
				db: db,
			}
			got, err := d.Has(context.Background(), tt.argument)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want.expect, got)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
