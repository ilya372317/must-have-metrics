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

// ShowAlertDTO DTO for represent request body and response body for show alert.
type ShowAlertDTO struct {
	Type string `valid:"in(gauge|counter)"`
	Name string `valid:"type(string)"`
}

// CreateShowAlertDTOFromRequest create ShowAlertDTO from given request.
func CreateShowAlertDTOFromRequest(request *http.Request) ShowAlertDTO {
	typ := chi.URLParam(request, typeURLParameter)
	name := chi.URLParam(request, nameURLParameter)

	return ShowAlertDTO{
		Type: typ,
		Name: name,
	}
}

// CreateShowAlertDTOFromMetrics create ShowAlertDTO from given Metrics DTO.
func CreateShowAlertDTOFromMetrics(metrics Metrics) ShowAlertDTO {
	return ShowAlertDTO{
		Type: metrics.MType,
		Name: metrics.ID,
	}
}

// Validate perform validation on ShowAlertDTO.
func (dto *ShowAlertDTO) Validate() (bool, error) {
	isValid, err := validator.ValidateRequired(*dto)
	if err != nil {
		err = fmt.Errorf("show dto is invalid: %w", err)
	}
	return isValid, err
}
