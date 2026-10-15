package helper

import (
	"fmt"
	"strings"

	"forum_backend/custom_err"
)

func IsEmptyText(str string) (string, bool) {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return "", true
	}
	return str, false
}

func ValidateValue(value int) error {
	if value != 1 && value != -1 {
		return fmt.Errorf("%w: like must be 1 or -1", custom_err.ErrInvalidInput)
	}
	return nil
}

func ValidateUsername(username *string) error {
	*username = strings.TrimSpace(*username)
	switch n := len(*username); {
	case n == 0:
		return custom_err.ErrEmptyUsername
	case n < 3:
		return custom_err.ErrUsernameTooShort
	case n > 30:
		return custom_err.ErrUsernameTooLong
	}

	for _, c := range *username {
		if !isAlphanumeric(c) && c != '_' {
			return custom_err.ErrInvalidUsername
		}
	}

	return nil
}

func ValidateName(name *string) error {
	*name = strings.TrimSpace(*name)
	if len(*name) > 32 {
		return custom_err.ErrNameTooLong
	}

	return nil
}

func ValidateDescription(desc *string) error {
	*desc = strings.TrimSpace(*desc)
	if len(*desc) > 500 {
		return custom_err.ErrDescTooLong
	}

	return nil
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
