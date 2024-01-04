package utils

import "fmt"

func Must1[T any](val T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("must not be error, but got: %v", err))
	}
	return val
}
