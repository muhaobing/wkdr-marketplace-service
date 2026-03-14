package utils

import "strings"

func GenKey(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}
