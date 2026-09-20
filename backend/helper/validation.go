package helper

import "strings"

func IsEmptyText(str string) (string, bool) {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return "", true
	}
	return str, false
}
