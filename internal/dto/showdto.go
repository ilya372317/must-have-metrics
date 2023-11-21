package dto

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/dto/validator"
)

type ShowAlertDTO struct {
	Type string `valid:"in(gauge|counter)"`
	Name string `valid:"-"`
}

func CreateShowAlertDTOFromRequest(request *http.Request) ShowAlertDTO {
	typ := chi.URLParam(request, "type")
	name := chi.URLParam(request, "name")

	return ShowAlertDTO{
		Type: typ,
		Name: name,
	}
}

func CreateShowAlertDTOFromMetrics(metrics Metrics) ShowAlertDTO {
	return ShowAlertDTO{
		Type: metrics.MType,
		Name: metrics.ID,
	}
}

func (dto *ShowAlertDTO) Validate() (bool, error) {
	nameNotEmpty := NotEmpty(dto.Name)
	typeNotEmpty := NotEmpty(dto.Type)
	if !nameNotEmpty || !typeNotEmpty {
		return false, errors.New("missing some required fields")
	}
	return validator.Validate(*dto)
}
