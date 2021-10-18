package wizard

import (
	"fmt"

	"github.com/mavolin/adam/pkg/errors"
)

type RetryError struct {
	desc string
}

func NewRetryError(description string) *RetryError {
	return &RetryError{desc: description}
}

func NewRetryErrorf(description string, a ...interface{}) *RetryError {
	return NewRetryError(fmt.Sprintf(description, a...))
}

func (r *RetryError) As(target interface{}) bool {
	switch target := target.(type) {
	case **errors.UserError:
		*target = errors.NewUserError(r.desc + "\n\nTry again!")
		return true
	case *errors.Error:
		*target = errors.NewUserError(r.desc + "\n\nTry again!")
		return true
	default:
		return false
	}
}

func (r *RetryError) Error() string {
	return "RetryError: " + r.desc
}
