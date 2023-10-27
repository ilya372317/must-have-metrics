package repository

import "github.com/ilya372317/must-have-metrics/internal/entity"

type AlertStorage interface {
	AddAlert(typ, name, data string) error
	GetAlert(name string) entity.Alert
}
