package dto

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/server/dto/validator"
)

type UpdateAlertDTO struct {
	Type string `valid:"in(gauge|counter)"`
	Name string `valid:"-"`
	Data string `valid:"stringisnumber"`
}

func CreateAlertDTOFromRequest(request *http.Request) UpdateAlertDTO {
	typ := chi.URLParam(request, "type")
	name := chi.URLParam(request, "name")
	value := chi.URLParam(request, "value")

	return UpdateAlertDTO{
		Type: typ,
		Name: name,
		Data: value,
	}
}

func (dto *UpdateAlertDTO) Validate() (bool, error) {
	if result := notEmpty(dto.Name); !result {
		return result, errors.New("name of alert not able to be empty")
	}
	return validator.Validate(*dto)
}
