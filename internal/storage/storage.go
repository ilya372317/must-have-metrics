package storage

import (
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type AlertStorage interface {
	SetAlert(name string, alert entity.Alert)
	UpdateAlert(name string, alert entity.Alert) error
	GetAlert(name string) (entity.Alert, error)
	HasAlert(name string) bool
}
