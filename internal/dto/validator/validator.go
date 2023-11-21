package validator

import (
	"strconv"

	"github.com/asaskevich/govalidator"
)

func Validate(dataTransferObject interface{}) (bool, error) {
	registerCustomValidators()
	govalidator.SetFieldsRequiredByDefault(true)
	result, err := govalidator.ValidateStruct(dataTransferObject)
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
