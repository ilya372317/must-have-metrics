package dto

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/server/dto/validator"
)

type ShowAlertDTO struct {
	Type string `valid:"in(gauge|counter)"`
	Name string `valid:"-"`
}

func ShowAlertDTOFromRequest(request *http.Request) ShowAlertDTO {
	typ := chi.URLParam(request, "type")
	name := chi.URLParam(request, "name")

	return ShowAlertDTO{
		Type: typ,
		Name: name,
	}
}

func (dto *ShowAlertDTO) Validate() (bool, error) {
	if result := notEmpty(dto.Name); !result {
		return result, errors.New("name of alert not able to be empty")
	}
	return validator.Validate(*dto)
}
