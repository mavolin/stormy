package action

import (
	"fmt"

	"github.com/mavolin/adam/pkg/errors"
)

// =============================================================================
// EnabledError
// =====================================================================================

// EnabledError is the error returned if a user attempts to enable an Action
// that is already enabled.
//
// It makes itself available as a *errors.UserError when calling As.
type EnabledError struct {
	// Name is the name of the Action.
	Name string
}

var _ error = new(EnabledError)

func NewEnabledError(actionName string) *EnabledError {
	return &EnabledError{Name: actionName}
}

func (e *EnabledError) Error() string {
	return fmt.Sprintf("action: the %s action is already enabled in the invoking channel", e.Name)
}

func (e *EnabledError) As(target interface{}) bool {
	switch err := target.(type) {
	case **errors.UserError:
		*err = e.AsUserError()
		return true
	case *errors.Error:
		*err = e.AsUserError()
		return true
	default:
		return false
	}
}

func (e *EnabledError) AsUserError() *errors.UserError {
	return errors.NewUserErrorf("`%s` is already enabled in this channel.", e.Name)
}

// =============================================================================
// DisabledError
// =====================================================================================

// DisabledError is the error returned if a user attempts to disable an Action
// that is already disabled.
//
// It makes itself available as a *errors.UserError when calling As.
type DisabledError struct {
	// Name is the name of the Action.
	Name string
}

var _ error = new(DisabledError)

func NewDisabledError(actionName string) *DisabledError {
	return &DisabledError{Name: actionName}
}

func (e *DisabledError) Error() string {
	return fmt.Sprintf("action: the %s action is already disabled in the invoking channel", e.Name)
}

func (e *DisabledError) As(target interface{}) bool {
	switch err := target.(type) {
	case **errors.UserError:
		*err = e.AsUserError()
		return true
	case *errors.Error:
		*err = e.AsUserError()
		return true
	default:
		return false
	}
}

func (e *DisabledError) AsUserError() *errors.UserError {
	return errors.NewUserErrorf("`%s` is already disabled in this channel.", e.Name)
}
