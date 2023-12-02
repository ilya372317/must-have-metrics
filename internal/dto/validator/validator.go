package validator

import (
	"fmt"
	"strconv"

	"github.com/asaskevich/govalidator"
)

func Validate(dataTransferObject interface{}, allFieldReq bool) (bool, error) {
	govalidator.TagMap["stringisnumber"] = stringIsNumber
	govalidator.SetFieldsRequiredByDefault(allFieldReq)
	result, err := govalidator.ValidateStruct(dataTransferObject)
	if err != nil {
		err = fmt.Errorf("sturct is invalid: %w", err)
	}
	return result, err
}

func ValidateRequired(dataTransferObject interface{}) (bool, error) {
	return Validate(dataTransferObject, true)
}

func stringIsNumber(str string) bool {
	_, intErr := strconv.Atoi(str)
	_, floatErr := strconv.ParseFloat(str, 64)
	return intErr == nil || floatErr == nil
}
