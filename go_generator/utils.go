package go_generator

func MapItems[K comparable, V interface{}](mapObj map[K]V) []V {
	var values []V
	for _, v := range mapObj {
		values = append(values, v)
	}

	return values
}

func Map[T1 interface{}, T2 interface{}](arr []T1, f func(T1) T2) []T2 {
	var values []T2
	for _, v := range arr {
		values = append(values, f(v))
	}

	return values
}

func Flatten[T interface{}](matrix [][]T) []T {
	var values []T
	for _, v := range matrix {
		values = append(values, v...)
	}

	return values
}
