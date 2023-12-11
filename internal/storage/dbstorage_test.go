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
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var db *sql.DB

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

		defer rows.Close()

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

func floatPointer(f float64) *float64 {
	return &f
}

func intPointer(i int64) *int64 {
	return &i
}

func setupDatabase(t *testing.T) (*dockertest.Pool, *dockertest.Resource) {
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
