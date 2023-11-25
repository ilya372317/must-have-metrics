package validator

import (
	"fmt"
	"strconv"

	"github.com/asaskevich/govalidator"
)

func Validate(dataTransferObject interface{}) (bool, error) {
	registerCustomValidators()
	govalidator.SetFieldsRequiredByDefault(true)
	result, err := govalidator.ValidateStruct(dataTransferObject)
	if err != nil {
		err = fmt.Errorf("sturct is invalid: %w", err)
	}
	return result, err
}

func registerCustomValidators() {
	govalidator.TagMap["stringisnumber"] = stringIsNumber
}

func stringIsNumber(str string) bool {
	_, intErr := strconv.Atoi(str)
	_, floatErr := strconv.ParseFloat(str, 64)
	return intErr == nil || floatErr == nil
}
