package storage

import (
	"context"
	"fmt"
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

func TestDatabaseStorage_All(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(mock sqlmock.Sqlmock)
		want    []entity.Alert
		wantErr bool
	}{
		{
			name: "Success - retrieve multiple alerts",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "type", "float_value", "int_value"}).
					AddRow("alert1", "gauge", nil, nil).
					AddRow("alert2", "counter", nil, nil)
				mock.ExpectQuery(regexp.QuoteMeta(selectAllMetricsQuery)).WillReturnRows(rows)
			},
			want: []entity.Alert{
				{Name: "alert1", Type: "gauge", FloatValue: nil, IntValue: nil},
				{Name: "alert2", Type: "counter", FloatValue: nil, IntValue: nil},
			},
			wantErr: false,
		},
		{
			name: "Success - no rows found",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "type", "float_value", "int_value"})
				mock.ExpectQuery(regexp.QuoteMeta(selectAllMetricsQuery)).WillReturnRows(rows)
			},
			want:    []entity.Alert{},
			wantErr: false,
		},
		{
			name: "Failure - scan error",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "type", "float_value", "int_value"}).
					AddRow("alert1", "gauge", "invalid_float", 10)
				mock.ExpectQuery(regexp.QuoteMeta(selectAllMetricsQuery)).WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() {
				_ = db.Close()
			}()

			tt.mock(mock)

			d := DatabaseStorage{db: db}
			got, err := d.All(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseStorage_AllWithKeys(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(mock sqlmock.Sqlmock)
		want    map[string]entity.Alert
		wantErr bool
	}{
		{
			name: "Success - retrieve multiple alerts",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "type", "float_value", "int_value"}).
					AddRow("alert1", "gauge", 1.23, nil).
					AddRow("alert2", "counter", nil, 10)
				mock.ExpectQuery(regexp.QuoteMeta(selectAllMetricsQuery)).WillReturnRows(rows)
			},
			want: map[string]entity.Alert{
				"alert1": {Name: "alert1", Type: "gauge", FloatValue: floatPointer(1.23), IntValue: nil},
				"alert2": {Name: "alert2", Type: "counter", FloatValue: nil, IntValue: intPointer(10)},
			},
			wantErr: false,
		},
		{
			name: "Success - no rows found",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "type", "float_value", "int_value"})
				mock.ExpectQuery(regexp.QuoteMeta(selectAllMetricsQuery)).WillReturnRows(rows)
			},
			want:    map[string]entity.Alert{},
			wantErr: false,
		},
		{
			name: "Failure - scan error",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "type", "float_value", "int_value"}).
					AddRow("alert1", "gauge", "invalid_float", 10)
				mock.ExpectQuery(regexp.QuoteMeta(selectAllMetricsQuery)).WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() {
				_ = db.Close()
			}()

			tt.mock(mock)

			d := DatabaseStorage{db: db}
			got, err := d.AllWithKeys(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseStorage_Fill(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(mock sqlmock.Sqlmock, m map[string]entity.Alert)
		input   map[string]entity.Alert
		wantErr bool
	}{
		{
			name: "Success - fill database with alerts",
			mock: func(mock sqlmock.Sqlmock, m map[string]entity.Alert) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metrics`)).WillReturnResult(sqlmock.NewResult(0, 1))
				for id, alert := range m {
					mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metrics ("id", "type", "float_value", "int_value") VALUES ($1, $2, $3, $4)`)).WithArgs(id, alert.Type, alert.FloatValue, alert.IntValue).WillReturnResult(sqlmock.NewResult(1, 1))
				}
				mock.ExpectCommit()
			},
			input: map[string]entity.Alert{
				"alert1": {Name: "alert1", Type: "gauge", FloatValue: floatPointer(1.23), IntValue: nil},
				"alert2": {Name: "alert2", Type: "counter", FloatValue: nil, IntValue: intPointer(10)},
			},
			wantErr: false,
		},
		{
			name: "Failure - begin transaction error",
			mock: func(mock sqlmock.Sqlmock, m map[string]entity.Alert) {
				mock.ExpectBegin().WillReturnError(fmt.Errorf("begin transaction error"))
			},
			input:   map[string]entity.Alert{},
			wantErr: true,
		},
		{
			name: "Failure - delete existing records error",
			mock: func(mock sqlmock.Sqlmock, m map[string]entity.Alert) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metrics`)).WillReturnError(fmt.Errorf("delete error"))
				mock.ExpectRollback()
			},
			input:   map[string]entity.Alert{},
			wantErr: true,
		},
		{
			name: "Failure - insert alert error",
			mock: func(mock sqlmock.Sqlmock, m map[string]entity.Alert) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metrics`)).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metrics ("id", "type", "float_value", "int_value") VALUES ($1, $2, $3, $4)`)).WillReturnError(fmt.Errorf("insert error"))
				mock.ExpectRollback()
			},
			input: map[string]entity.Alert{
				"alert1": {Name: "alert1", Type: "gauge", FloatValue: floatPointer(1.23), IntValue: nil},
			},
			wantErr: true,
		},
		{
			name: "Failure - commit transaction error",
			mock: func(mock sqlmock.Sqlmock, m map[string]entity.Alert) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metrics`)).WillReturnResult(sqlmock.NewResult(0, 1))
				for id, alert := range m {
					mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metrics ("id", "type", "float_value", "int_value") VALUES ($1, $2, $3, $4)`)).WithArgs(id, alert.Type, alert.FloatValue, alert.IntValue).WillReturnResult(sqlmock.NewResult(1, 1))
				}
				mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))
			},
			input: map[string]entity.Alert{
				"alert1": {Name: "alert1", Type: "gauge", FloatValue: floatPointer(1.23), IntValue: nil},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mock(mock, tt.input)

			d := DatabaseStorage{db: db}
			err = d.Fill(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func floatPointer(f float64) *float64 {
	return &f
}

func intPointer(i int64) *int64 {
	return &i
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
