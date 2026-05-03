package common

import (
	"errors"
	"fmt"
)

func SwitchError[T any](err error, mapping map[error]T) T {
	var zero T

	fmt.Printf("err is %v\n", err)
	for target, value := range mapping {
		if errors.Is(err, target) {
			return value
		}
	}

	return zero
}
