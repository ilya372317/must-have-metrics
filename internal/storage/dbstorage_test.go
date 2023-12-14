package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var db *sql.DB

func TestDatabaseStorage_Fill(t *testing.T) {
	pool, resource := setupDatabase(t)
	defer teardownDatabase(t, pool, resource)
	tests := []struct {
		name   string
		arg    map[string]entity.Alert
		fields []entity.Alert
	}{
		{
			name: "success case",
			arg: map[string]entity.Alert{
				"alert": {
					Type:       "gauge",
					Name:       "alert",
					FloatValue: floatPointer(1.234),
				},
				"alert1": {
					Type:     "counter",
					Name:     "alert1",
					IntValue: intPointer(12345),
				},
			},
			fields: []entity.Alert{
				{
					Type:       "gauge",
					Name:       "alert3",
					FloatValue: floatPointer(1.23456789),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillDatabase(t, tt.fields)
			dbStorage := &DatabaseStorage{DB: db}
			err := dbStorage.Fill(context.Background(), tt.arg)
			require.NoError(t, err)

			got, err := dbStorage.All(context.Background())
			require.NoError(t, err)

			want := make([]entity.Alert, 0, len(tt.arg))

			for _, alert := range tt.arg {
				want = append(want, alert)
			}

			assert.Equal(t, want, got)
			clearDatabase(t)
		})
	}
}

func TestDatabaseStorage_AllWithKeys(t *testing.T) {
	pool, resource := setupDatabase(t)
	defer teardownDatabase(t, pool, resource)
	fields := []entity.Alert{
		{
			Type:       "gauge",
			Name:       "alert1",
			FloatValue: floatPointer(1.234),
		},
		{
			Type:     "counter",
			Name:     "alert2",
			IntValue: intPointer(1234),
		},
		{
			Type:       "gauge",
			Name:       "alert3",
			FloatValue: floatPointer(1.234234253),
		},
	}
	want := make(map[string]entity.Alert)
	for _, field := range fields {
		want[field.Name] = field
	}
	tests := []struct {
		name   string
		fields []entity.Alert
		want   map[string]entity.Alert
	}{
		{
			name:   "success filled case",
			fields: fields,
			want:   want,
		},
		{
			name:   "success empty storage case",
			fields: nil,
			want:   map[string]entity.Alert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillDatabase(t, tt.fields)
			dbStorage := &DatabaseStorage{DB: db}
			got, err := dbStorage.AllWithKeys(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			clearDatabase(t)
		})
	}
}

func TestDatabaseStorage_Has(t *testing.T) {
	pool, resource := setupDatabase(t)
	defer teardownDatabase(t, pool, resource)
	tests := []struct {
		name   string
		arg    string
		fileds []entity.Alert
		want   bool
	}{
		{
			name: "success simple case",
			arg:  "alert1",
			fileds: []entity.Alert{
				{
					Type:       "gauge",
					Name:       "alert1",
					FloatValue: floatPointer(1.234),
				},
				{
					Type:     "counter",
					Name:     "alert2",
					IntValue: intPointer(1234),
				},
			},
			want: true,
		},
		{
			name: "negative simple case",
			arg:  "alert",
			fileds: []entity.Alert{
				{
					Type:       "gauge",
					Name:       "alert1",
					FloatValue: floatPointer(1.234),
				},
				{
					Type:     "counter",
					Name:     "alert2",
					IntValue: intPointer(1234),
				},
			},
			want: false,
		},
		{
			name:   "simple negative case with empty storage",
			arg:    "alert",
			fileds: nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillDatabase(t, tt.fileds)
			dbStorage := &DatabaseStorage{DB: db}

			got, err := dbStorage.Has(context.Background(), tt.arg)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)

			clearDatabase(t)
		})
	}
}

func TestDatabaseStorage_All(t *testing.T) {
	pool, resource := setupDatabase(t)
	defer teardownDatabase(t, pool, resource)
	tests := []struct {
		name   string
		fileds []entity.Alert
		want   []entity.Alert
	}{
		{
			name: "success simple case",
			fileds: []entity.Alert{
				{
					Type:       "gauge",
					Name:       "alert1",
					FloatValue: floatPointer(1.234),
				},
				{
					Type:     "counter",
					Name:     "alert2",
					IntValue: intPointer(1234),
				},
				{
					Type:       "gauge",
					Name:       "alert3",
					FloatValue: floatPointer(1.234234253),
				},
			},
			want: []entity.Alert{
				{
					Type:       "gauge",
					Name:       "alert1",
					FloatValue: floatPointer(1.234),
				},
				{
					Type:     "counter",
					Name:     "alert2",
					IntValue: intPointer(1234),
				},
				{
					Type:       "gauge",
					Name:       "alert3",
					FloatValue: floatPointer(1.234234253),
				},
			},
		},
		{
			name:   "success case with empty storage",
			fileds: nil,
			want:   []entity.Alert{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillDatabase(t, tt.fileds)
			dbStorage := &DatabaseStorage{DB: db}
			got, err := dbStorage.All(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			clearDatabase(t)
		})
	}
}

func TestSave(t *testing.T) {
	pool, resource := setupDatabase(t)
	defer teardownDatabase(t, pool, resource)

	t.Run("Save", func(t *testing.T) {
		storage := DatabaseStorage{DB: db}

		alert := entity.Alert{
			Name:       "test1",
			Type:       "gauge",
			FloatValue: floatPointer(1.23),
		}

		err := storage.Save(context.Background(), alert.Name, alert)
		require.NoError(t, err)

		var savedAlert entity.Alert
		err = db.QueryRowContext(context.Background(),
			`SELECT "id", "type", "float_value", "int_value" FROM metrics WHERE "id" = $1`,
			alert.Name).Scan(&savedAlert.Name, &savedAlert.Type, &savedAlert.FloatValue, &savedAlert.IntValue)

		require.NoError(t, err)

		assert.Equal(t, alert, savedAlert)
	})
}

func TestDatabaseStorage_Get(t *testing.T) {
	pool, resource := setupDatabase(t)
	defer teardownDatabase(t, pool, resource)
	tests := []struct {
		name    string
		wantErr bool
		want    entity.Alert
		fields  []entity.Alert
		arg     string
	}{
		{
			name:    "success simple case",
			wantErr: false,
			want: entity.Alert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: floatPointer(1.234),
			},
			fields: []entity.Alert{
				{
					Type:       "gauge",
					Name:       "alert",
					FloatValue: floatPointer(1.234),
				},
			},
			arg: "alert",
		},
		{
			name:    "simple error case",
			wantErr: true,
			want:    entity.Alert{},
			fields:  nil,
			arg:     "alert",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillDatabase(t, tt.fields)
			dbStorage := &DatabaseStorage{DB: db}
			got, err := dbStorage.Get(context.Background(), tt.arg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)

			clearDatabase(t)
		})
	}
}

func TestBulkInsertOrUpdate(t *testing.T) {
	pool, resource := setupDatabase(t)
	defer teardownDatabase(t, pool, resource)

	t.Run("BulkInsertOrUpdate", func(t *testing.T) {
		storage := DatabaseStorage{DB: db}

		_, err := db.ExecContext(context.Background(),
			`INSERT INTO metrics (id, type, float_value, int_value) 
		VALUES ($1,$2,$3,$4)`,
			"test1", "counter", nil, intPointer(123))
		require.NoError(t, err)

		alerts := []entity.Alert{
			{Name: "test1", Type: "gauge", FloatValue: floatPointer(1.23)},
			{Name: "test2", Type: "counter", IntValue: intPointer(1)},
		}

		err = storage.BulkInsertOrUpdate(context.Background(), alerts)
		if err != nil {
			t.Errorf("BulkInsertOrUpdate failed: %s", err)
		}

		insertedAlerts := make([]entity.Alert, 0, len(alerts))

		rows, err := db.QueryContext(context.Background(),
			`SELECT "id", "type", "float_value", "int_value" FROM metrics`)

		defer func() {
			err = rows.Close()
			require.NoError(t, err)
		}()

		require.NoError(t, err)

		for rows.Next() {
			insertedAlert := entity.Alert{}
			err := rows.Scan(&insertedAlert.Name, &insertedAlert.Type, &insertedAlert.FloatValue, &insertedAlert.IntValue)
			require.NoError(t, err)
			insertedAlerts = append(insertedAlerts, insertedAlert)
		}

		require.NoError(t, rows.Err())

		assert.Equal(t, alerts, insertedAlerts)
	})
}

func TestDatabaseStorage_GetByIds(t *testing.T) {
	pool, resource := setupDatabase(t)
	defer teardownDatabase(t, pool, resource)
	tests := []struct {
		name     string
		fields   []entity.Alert
		argument []string
		want     []entity.Alert
	}{
		{
			name: "success case with filled storage",
			fields: []entity.Alert{
				{
					Type:       "gauge",
					Name:       "alert1",
					FloatValue: floatPointer(10.234),
					IntValue:   nil,
				},
				{
					Type:       "counter",
					Name:       "alert2",
					FloatValue: nil,
					IntValue:   intPointer(1234),
				},
				{
					Type:       "gauge",
					Name:       "alert3",
					FloatValue: floatPointer(23.34),
					IntValue:   nil,
				},
			},
			argument: []string{"alert1", "alert2", "alert3"},
			want: []entity.Alert{
				{
					Type:       "gauge",
					Name:       "alert1",
					FloatValue: floatPointer(10.234),
					IntValue:   nil,
				},
				{
					Type:       "counter",
					Name:       "alert2",
					FloatValue: nil,
					IntValue:   intPointer(1234),
				},
				{
					Type:       "gauge",
					Name:       "alert3",
					FloatValue: floatPointer(23.34),
					IntValue:   nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			databaseStorage := DatabaseStorage{DB: db}
			fillDatabase(t, tt.fields)
			got, err := databaseStorage.GetByIDs(context.Background(), tt.argument)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)

			clearDatabase(t)
		})
	}
}

func floatPointer(f float64) *float64 {
	return &f
}

func intPointer(i int64) *int64 {
	return &i
}

func clearDatabase(t *testing.T) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`DELETE FROM metrics`)
	require.NoError(t, err)
}

func setupDatabase(t *testing.T) (*dockertest.Pool, *dockertest.Resource) {
	err := logger.Init()
	require.NoError(t, err)
	t.Helper()
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=metrics_test"})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	port := resource.GetPort("5432/tcp")
	connectionString := fmt.Sprintf(
		"host=localhost port=%s user=postgres password=secret dbname=metrics_test sslmode=disable",
		port,
	)

	if err = pool.Retry(func() error {
		var err error
		db, err = sql.Open("pgx", connectionString)
		if err != nil {
			return err //nolint:wrapcheck
		}
		return db.Ping() //nolint:wrapcheck
	}); err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		t.Fatalf("failed create driver")
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../db/migrations",
		"metrics_test",
		driver,
	)
	if err != nil {
		db.Close()
		t.Fatalf("Migration failed: %s", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		db.Close()
		t.Fatalf("Migration failed: %s", err)
	}

	return pool, resource
}

func teardownDatabase(t *testing.T, pool *dockertest.Pool, resource *dockertest.Resource) {
	t.Helper()
	_ = db.Close()
	if err := pool.Purge(resource); err != nil {
		t.Fatalf("Could not purge resource: %s", err)
	}
}

func fillDatabase(t *testing.T, fields []entity.Alert) {
	t.Helper()
	for _, existedAlert := range fields {
		_, err := db.ExecContext(context.Background(),
			`INSERT INTO metrics ("id", "type", "float_value", "int_value") 
				VALUES ($1, $2, $3, $4)`, existedAlert.Name, existedAlert.Type,
			existedAlert.FloatValue, existedAlert.IntValue)
		require.NoError(t, err)
	}
}
