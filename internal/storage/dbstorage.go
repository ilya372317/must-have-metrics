package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	failedMakeRollbackErrPattern = "failed make rollback: %v"
	failedCloseRowsErrPattern    = "failed close rows connection: %v"
	selectAllMetricsQuery        = `SELECT "id", "type", "float_value", "int_value" FROM metrics`
	failedExecuteQueryErrPattern = "failed to execute query: %w"
	failedScanRowErrPattern      = "failed to scan row: %w"
	iterationInRowsErrPattern    = "error iterating through rows: %w"
	failedRollbackErrPattern     = "failed rollback: %v"
	stepForDelayRetry            = 2
)

type DatabaseStorage struct {
	DB *sql.DB
}

func (d *DatabaseStorage) Save(ctx context.Context, name string, alert entity.Alert) error {
	operation := func() error {
		_, err := d.DB.ExecContext(ctx,
			`INSERT INTO metrics ("id", "type", "int_value", "float_value") VALUES ($1,$2,$3,$4)`,
			name, alert.Type, alert.IntValue, alert.FloatValue)
		if err != nil {
			return fmt.Errorf("failed make insert request: %w", err)
		}
		return nil
	}

	return withRetries(operation)
}

func (d *DatabaseStorage) Update(ctx context.Context, name string, alert entity.Alert) error {
	operation := func() error {
		_, err := d.DB.ExecContext(ctx,
			`UPDATE metrics SET "type" = $1, "float_value" = $2, "int_value" = $3 WHERE id = $4`,
			alert.Type, alert.FloatValue, alert.IntValue, name)
		if err != nil {
			return fmt.Errorf("failed update alert in database: %w", err)
		}
		return nil
	}
	return withRetries(operation)
}

func (d *DatabaseStorage) Get(ctx context.Context, name string) (entity.Alert, error) {
	resultAlert := entity.Alert{}
	operation := func() error {
		var floatValue sql.NullFloat64
		var intValue sql.NullInt64
		row := d.DB.QueryRowContext(ctx,
			`SELECT "id", "type", "float_value", "int_value" FROM metrics WHERE id = $1`,
			name)
		err := row.Scan(&resultAlert.Name, &resultAlert.Type, &floatValue, &intValue)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("alert with id '%s' not found: %w", name, err)
			}
			return fmt.Errorf("failed to get alert with id '%s' from database: %w", name, err)
		}

		if floatValue.Valid {
			resultAlert.FloatValue = &floatValue.Float64
		}
		if intValue.Valid {
			resultAlert.IntValue = &intValue.Int64
		}
		return nil
	}

	err := withRetries(operation)
	if err != nil {
		return entity.Alert{}, err
	}

	return resultAlert, nil
}

func (d *DatabaseStorage) Has(ctx context.Context, name string) (bool, error) {
	var id string
	var result bool

	operation := func() error {
		err := d.DB.QueryRowContext(ctx,
			`SELECT "id" FROM metrics WHERE id = $1`, name).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				result = false
				return nil
			}
			result = false
			return fmt.Errorf("failed make has query: %w", err)
		}
		result = true
		return nil
	}

	err := withRetries(operation)
	if err != nil {
		return false, err
	}

	return result, nil
}

func (d *DatabaseStorage) All(ctx context.Context) ([]entity.Alert, error) {
	alerts := make([]entity.Alert, 0, 100) //nolint:nolintlint,gomnd

	operation := func() error {
		rows, err := d.DB.QueryContext(ctx,
			selectAllMetricsQuery)
		defer func() {
			if err = rows.Close(); err != nil {
				logger.Log.Warnf(failedCloseRowsErrPattern, err)
			}
		}()
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("failed get all records from database: %w", err)
		}

		for rows.Next() {
			newAlert := entity.Alert{}
			err = rows.Scan(&newAlert.Name, &newAlert.Type, &newAlert.FloatValue, &newAlert.IntValue)
			if err != nil {
				return fmt.Errorf("failed scan data in all query:%w ", err)
			}

			alerts = append(alerts, newAlert)
		}

		if rows.Err() != nil {
			return fmt.Errorf("something went wrong on scan rows: %w", err)
		}
		return nil
	}

	if err := withRetries(operation); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (d *DatabaseStorage) AllWithKeys(ctx context.Context) (map[string]entity.Alert, error) {
	alerts := make(map[string]entity.Alert)
	operation := func() error {
		rows, err := d.DB.QueryContext(ctx, selectAllMetricsQuery)
		if err != nil {
			return fmt.Errorf(failedExecuteQueryErrPattern, err)
		}
		defer func() {
			_ = rows.Close()
		}()

		for rows.Next() {
			var alert entity.Alert
			if err = rows.Scan(&alert.Name, &alert.Type, &alert.FloatValue, &alert.IntValue); err != nil {
				return fmt.Errorf(failedScanRowErrPattern, err)
			}
			alerts[alert.Name] = alert
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf(iterationInRowsErrPattern, err)
		}
		return nil
	}

	if err := withRetries(operation); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (d *DatabaseStorage) Fill(ctx context.Context, m map[string]entity.Alert) error {
	operation := func() error {
		tx, err := d.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM metrics`)
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				logger.Log.Warnf(failedMakeRollbackErrPattern, err)
			}
			return fmt.Errorf("failed to delete existing records: %w", err)
		}

		for id, alert := range m {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO metrics ("id", "type", "float_value", "int_value") VALUES ($1, $2, $3, $4)`,
				id, alert.Type, alert.FloatValue, alert.IntValue)
			if err != nil {
				if err = tx.Rollback(); err != nil {
					logger.Log.Warnf(failedMakeRollbackErrPattern, err)
				}
				return fmt.Errorf("failed to insert alert with id %s: %w", id, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	}

	return withRetries(operation)
}

func (d *DatabaseStorage) BulkInsertOrUpdate(ctx context.Context, alerts []entity.Alert) error {
	operation := func() error {
		tx, err := d.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed begin transaction on insert or update: %w", err)
		}

		preparedQuery, err := tx.PrepareContext(ctx, `INSERT INTO metrics ("id", "type", "float_value", "int_value") 
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT (id) 
	DO UPDATE SET "type" = excluded.type, "float_value" = excluded.float_value, "int_value" = excluded.int_value`)

		if err != nil {
			if err = tx.Rollback(); err != nil {
				logger.Log.Warnf(failedRollbackErrPattern, err)
			}
			return fmt.Errorf("failed prepare query on update or insert: %w", err)
		}
		defer func() {
			_ = preparedQuery.Close()
		}()

		for _, alert := range alerts {
			_, err = preparedQuery.ExecContext(ctx, alert.Name, alert.Type, alert.FloatValue, alert.IntValue)
			if err != nil {
				if err = tx.Rollback(); err != nil {
					logger.Log.Warnf(failedRollbackErrPattern, err)
				}
				return fmt.Errorf("failed insert alert %v: %w", alert, err)
			}
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed commit changes on update or insert: %w", err)
		}

		return nil
	}
	return withRetries(operation)
}

func (d *DatabaseStorage) GetByIDs(ctx context.Context, ids []string) ([]entity.Alert, error) {
	if len(ids) == 0 {
		return []entity.Alert{}, nil
	}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholderStr := strings.Join(placeholders, ",")

	query := fmt.Sprintf(
		`SELECT "id", "type", "float_value", "int_value" FROM metrics WHERE "id" IN (%s)`,
		placeholderStr)

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := d.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(failedExecuteQueryErrPattern, err)
	}
	defer func() {
		err = rows.Close()
		logger.Log.Warnf("failed close rows: %v", err)
	}()

	var alerts []entity.Alert
	for rows.Next() {
		var alert entity.Alert
		var floatValue sql.NullFloat64
		var intValue sql.NullInt64

		if err := rows.Scan(&alert.Name, &alert.Type, &floatValue, &intValue); err != nil {
			return nil, fmt.Errorf(failedScanRowErrPattern, err)
		}

		if floatValue.Valid {
			alert.FloatValue = &floatValue.Float64
		}
		if intValue.Valid {
			alert.IntValue = &intValue.Int64
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(iterationInRowsErrPattern, err)
	}

	return alerts, nil
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
	return false
}

func withRetries(fn func() error) error {
	const maxRetries = 3
	delay := time.Second
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !isRetriableError(err) {
			return err
		}
		time.Sleep(delay)
		delay += time.Second * stepForDelayRetry
	}
	return err
}
