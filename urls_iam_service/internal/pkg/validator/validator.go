package validator

import "github.com/go-playground/validator/v10"

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) Validate(data interface{}) error {
	return v.validate.Struct(data)
}

func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}
