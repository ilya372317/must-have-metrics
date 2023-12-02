package dto

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/dto/validator"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type UpdateAlertDTO struct {
	Type string `valid:"in(gauge|counter)"`
	Name string `valid:"type(string)"`
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
	if metrics.Delta != nil && metrics.MType == entity.TypeCounter {
		value = fmt.Sprintf("%d", *metrics.Delta)
	} else if metrics.Value != nil && metrics.MType == entity.TypeGauge {
		value = strconv.FormatFloat(*metrics.Value, 'f', -1, 64)
	}
	return UpdateAlertDTO{
		Type: metrics.MType,
		Name: metrics.ID,
		Data: value,
	}
}

func (dto *UpdateAlertDTO) Validate() (bool, error) {
	switch dto.Type {
	case entity.TypeGauge:
		if !stringIsFloat(dto.Data) {
			return false, errors.New("invalid value. value must be float")
		}
	case entity.TypeCounter:
		if !stringIsInt(dto.Data) {
			return false, errors.New("invalid value. value must be int")
		}
	}

	isValid, err := validator.ValidateRequired(*dto)
	if err != nil {
		err = fmt.Errorf("update dto is invalid: %w", err)
	}

	return isValid, err
}

func stringIsFloat(str string) bool {
	_, floatErr := strconv.ParseFloat(str, 64)
	return floatErr == nil
}
func stringIsInt(str string) bool {
	_, intErr := strconv.Atoi(str)
	return intErr == nil
}
