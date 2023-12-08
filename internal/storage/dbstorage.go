package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

const (
	failedMakeRollbackErrPattern = "failed make rollback: %v"
	failedCloseRowsErrPattern    = "failed close rows connection: %v"
	selectAllMetricsQuery        = `SELECT "id", "type", "float_value", "int_value" FROM metrics`
)

type DatabaseStorage struct {
	db *sql.DB
}

func (d *DatabaseStorage) Save(ctx context.Context, name string, alert entity.Alert) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO metrics ("id", "type", "int_value", "float_value") VALUES ($1,$2,$3,$4)`,
		name, alert.Type, alert.IntValue, alert.FloatValue)
	if err != nil {
		return fmt.Errorf("failed make insert request: %w", err)
	}
	return nil
}

func (d *DatabaseStorage) Update(ctx context.Context, name string, alert entity.Alert) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE metrics SET "type" = $1, "float_value" = $2, "int_value" = $3 WHERE id = $4`,
		alert.Type, alert.FloatValue, alert.IntValue, name)
	if err != nil {
		return fmt.Errorf("failed update alert in database: %w", err)
	}
	return nil
}

func (d *DatabaseStorage) Get(ctx context.Context, name string) (entity.Alert, error) {
	resultAlert := entity.Alert{}
	var floatValue sql.NullFloat64
	var intValue sql.NullInt64
	row := d.db.QueryRowContext(ctx,
		`SELECT "id", "type", "float_value", "int_value" FROM metrics WHERE id = $1`,
		name)
	err := row.Scan(&resultAlert.Name, &resultAlert.Type, &floatValue, &intValue)
	if err != nil {
		return entity.Alert{}, fmt.Errorf("failed get alert from database: %w", err)
	}

	if floatValue.Valid {
		value := floatValue.Float64
		resultAlert.FloatValue = &value
	}
	if intValue.Valid {
		value := intValue.Int64
		resultAlert.IntValue = &value
	}
	return resultAlert, nil
}

func (d *DatabaseStorage) Has(ctx context.Context, name string) (bool, error) {
	var id string
	err := d.db.QueryRowContext(ctx,
		`SELECT "id" FROM metrics WHERE id = $1`, name).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed make has query: %w", err)
	}
	return true, nil
}

func (d *DatabaseStorage) All(ctx context.Context) ([]entity.Alert, error) {
	alerts := make([]entity.Alert, 0, 100)
	rows, err := d.db.QueryContext(ctx,
		selectAllMetricsQuery)
	defer func() {
		if err = rows.Close(); err != nil {
			logger.Get().Warnf(failedCloseRowsErrPattern, err)
		}
	}()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return alerts, nil
		}
		return nil, fmt.Errorf("failed get all records from database: %w", err)
	}

	for rows.Next() {
		newAlert := entity.Alert{}
		err = rows.Scan(&newAlert.Name, &newAlert.Type, &newAlert.FloatValue, &newAlert.IntValue)
		if err != nil {
			return nil, fmt.Errorf("failed scan data in all query:%w ", err)
		}

		alerts = append(alerts, newAlert)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed get all request: %w", err)
	}

	return alerts, nil
}

func (d *DatabaseStorage) AllWithKeys(ctx context.Context) (map[string]entity.Alert, error) {
	alerts := make(map[string]entity.Alert)
	rows, err := d.db.QueryContext(ctx, selectAllMetricsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var alert entity.Alert
		if err := rows.Scan(&alert.Name, &alert.Type, &alert.FloatValue, &alert.IntValue); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		alerts[alert.Name] = alert
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through rows: %w", err)
	}

	return alerts, nil
}

func (d *DatabaseStorage) Fill(ctx context.Context, m map[string]entity.Alert) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM metrics`)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			logger.Get().Warnf(failedMakeRollbackErrPattern, err)
		}
		return fmt.Errorf("failed to delete existing records: %w", err)
	}

	for id, alert := range m {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO metrics ("id", "type", "float_value", "int_value") VALUES ($1, $2, $3, $4)`,
			id, alert.Type, alert.FloatValue, alert.IntValue)
		if err != nil {
			if err = tx.Rollback(); err != nil {
				logger.Get().Warnf(failedMakeRollbackErrPattern, err)
			}
			return fmt.Errorf("failed to insert alert with id %s: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
