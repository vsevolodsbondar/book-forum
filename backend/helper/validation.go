package helper

import (
	"fmt"
	"forum_backend/custom_err"
	"strings"
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
