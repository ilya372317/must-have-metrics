package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type DatabaseStorage struct {
	db *sql.DB
}

func (d *DatabaseStorage) Save(name string, alert entity.Alert) {
	_, _ = d.db.ExecContext(context.TODO(),
		`INSERT INTO metrics ("id", "type", int_value, float_value) VALUES ($1,$2,$3,$4)`,
		name, alert.Type, alert.IntValue, alert.FloatValue)
	//TODO: return err
}

func (d *DatabaseStorage) Update(name string, alert entity.Alert) error {
	_, err := d.db.ExecContext(context.TODO(),
		`UPDATE metrics SET "type" = $1, "float_value" = $2, "int_value" = $3 WHERE id = $4`,
		alert.Type, alert.FloatValue, alert.IntValue, name)
	if err != nil {
		return fmt.Errorf("failed update alert in database: %w", err)
	}
	return nil
}

func (d *DatabaseStorage) Get(name string) (entity.Alert, error) {
	resultAlert := entity.Alert{}
	row := d.db.QueryRowContext(context.TODO(),
		`SELECT "id", "type", "float_value", "int_value" FROM metrics WHERE id = $1`,
		name)
	if row.Err() != nil {
		return entity.Alert{}, fmt.Errorf("invalid select query: %w", row.Err())
	}
	err := row.Scan(&resultAlert.Name, &resultAlert.Type, &resultAlert.FloatValue, resultAlert.IntValue)
	if err != nil {
		return entity.Alert{}, fmt.Errorf("failed get alert from database: %w", err)
	}
	return resultAlert, nil
}

func (d *DatabaseStorage) Has(name string) bool {
	rows, err := d.db.QueryContext(context.TODO(),
		`SELECT "id" FROM metrics`)
	defer rows.Close()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		//TODO: return err
	}
	if rows.Err() != nil {
		// TODO: return err
	}
	return true
}

func (d *DatabaseStorage) All() []entity.Alert {
	alerts := make([]entity.Alert, 0, 100)
	rows, err := d.db.QueryContext(context.TODO(),
		`SELECT "id", "type", "float_value", "int_value" FROM metrics`)
	defer rows.Close()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return alerts
		}
		// TODO: return err
	}

	for rows.Next() {
		newAlert := entity.Alert{}
		_ = rows.Scan(&newAlert.Name, &newAlert.Type, &newAlert.FloatValue, &newAlert.IntValue)
		// TODO: check err
		alerts = append(alerts, newAlert)
	}

	return alerts
}

func (d *DatabaseStorage) AllWithKeys() map[string]entity.Alert {
	// TODO implement me
	panic("implement me")
}

func (d *DatabaseStorage) Fill(m map[string]entity.Alert) {
	// TODO implement me
	panic("implement me")
}
