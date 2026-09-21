package handler

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewValidator() (*validator.Validate, error) {
	v := validator.New()

	if err := v.RegisterValidation("auth_email", validateEmail); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("auth_username", validateUsername); err != nil {
		return nil, err
	}

	return v, nil
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if len(email) > 254 {
		return false
	}

	for i := range len(email) {
		if email[i] > 127 {
			return false
		}
	}

	local, domain, ok := strings.Cut(email, "@")
	if !ok || len(local) > 64 {
		return false
	}

	return !strings.Contains(local, `"`) && !strings.HasPrefix(domain, "[")
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	for _, c := range username {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_') {
			return false
		}
	}

	return true
}
