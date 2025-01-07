package errutils

import (
	"errors"
	"fmt"
)

var Empty error = emptyErr{}

type emptyErr struct{}

func (e emptyErr) Error() string {
	return ""
}

func Newf(text string, a ...any) error {
	return errors.New(fmt.Sprintf(text, a...))
}
