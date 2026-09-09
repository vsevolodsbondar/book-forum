package helper

import "strings"

func IsEmptyText(str string) (string, bool) {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return "", false
	}
	return str, true
}
