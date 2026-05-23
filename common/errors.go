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

func SwitchErrorWithDefault[T any](err error, defaultValue T, mapping map[error]T) T {
	fmt.Printf("err is %v\n", err)
	for target, value := range mapping {
		if errors.Is(err, target) {
			return value
		}
	}

	return defaultValue
}

func SwitchErrorWithDefaultFunc[T any](err error, defaultValue func(err error) T, mapping map[error]T) T {
	fmt.Printf("err is %v\n", err)
	for target, value := range mapping {
		if errors.Is(err, target) {
			return value
		}
	}

	return defaultValue(err)
}

func WrapError(domainErr, rawError error) error {
	return fmt.Errorf("%w: %w", domainErr, rawError)
}
