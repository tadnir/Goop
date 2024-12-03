package utils

import (
	"iter"
	"strings"
)

func Map[T1 interface{}, T2 interface{}](arr iter.Seq[T1], f func(T1) T2) []T2 {
	var values []T2
	for v := range arr {
		values = append(values, f(v))
	}

	return values
}

func Capitalize(str string) string {
	return strings.ToUpper(string(str[0])) + str[1:]
}
