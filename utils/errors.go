package utils

import (
	// "errors"
	"fmt"
)

func NewError(code int, message string) error {
	return fmt.Errorf("%d: %s", code, message)
}

