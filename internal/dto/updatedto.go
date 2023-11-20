package dto

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/dto/validator"
)

type UpdateAlertDTO struct {
	Type string `valid:"in(gauge|counter)"`
	Name string `valid:"-"`
	Data string `valid:"stringisnumber"`
}

func CreateUpdateAlertDTOFromRequest(request *http.Request) UpdateAlertDTO {
	typ := chi.URLParam(request, "type")
	name := chi.URLParam(request, "name")
	value := chi.URLParam(request, "value")

	return UpdateAlertDTO{
		Type: typ,
		Name: name,
		Data: value,
	}
}

func CreateUpdateAlertDTOFromMetrics(metrics Metrics) UpdateAlertDTO {
	value := ""
	if metrics.Delta != nil {
		value = fmt.Sprintf("%d", *metrics.Delta)
	} else if metrics.Value != nil {
		value = fmt.Sprintf("%f", *metrics.Value)
	}
	return UpdateAlertDTO{
		Type: metrics.MType,
		Name: metrics.ID,
		Data: value,
	}
}

func (dto *UpdateAlertDTO) Validate() (bool, error) {
	if result := NotEmpty(dto.Name); !result {
		return result, errors.New("name of alert not able to be empty")
	}
	return validator.Validate(*dto)
}
