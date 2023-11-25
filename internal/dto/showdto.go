package dto

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/dto/validator"
)

const (
	typeURLParameter = "type"
	nameURLParameter = "name"
)

type ShowAlertDTO struct {
	Type string `valid:"in(gauge|counter)"`
	Name string `valid:"type(string)"`
}

func CreateShowAlertDTOFromRequest(request *http.Request) ShowAlertDTO {
	typ := chi.URLParam(request, typeURLParameter)
	name := chi.URLParam(request, nameURLParameter)

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
	isValid, err := validator.Validate(*dto)
	if err != nil {
		err = fmt.Errorf("show dto is invalid: %w", err)
	}
	return isValid, err
}
