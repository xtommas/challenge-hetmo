package validator

import (
	"github.com/go-playground/validator"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}
