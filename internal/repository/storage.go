package repository

import "github.com/ilya372317/must-have-metrics/internal/entity"

type AlertStorage interface {
	AddAlert(name string, alert entity.Alert)
	UpdateAlert(name string, alert entity.AlertValue) error
	GetAlert(name string) (entity.Alert, error)
	HasAlert(name string) bool
}
