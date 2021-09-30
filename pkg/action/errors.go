package action

import (
	"fmt"

	"github.com/mavolin/adam/pkg/errors"
)

// =============================================================================
// AlreadyEnabledError
// =====================================================================================

// AlreadyEnabledError is the error returned if a user attempts to enable an Action
// that is already enabled.
//
// It makes itself available as a *errors.UserError when calling As.
type AlreadyEnabledError struct {
	// Name is the name of the Action.
	Name string
}

var _ error = new(AlreadyEnabledError)

func NewAlreadyEnabledError(actionName string) *AlreadyEnabledError {
	return &AlreadyEnabledError{Name: actionName}
}

func (e *AlreadyEnabledError) Error() string {
	return fmt.Sprintf("action: the %s action is already enabled in the invoking channel", e.Name)
}

func (e *AlreadyEnabledError) As(target interface{}) bool {
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

func (e *AlreadyEnabledError) AsUserError() *errors.UserError {
	return errors.NewUserErrorf("`%s` is already enabled in this channel.", e.Name)
}

// =============================================================================
// ModifyDisabledActionError
// =====================================================================================

// ModifyDisabledActionError is the error returned if attempting to modify an
// action, that is currently disabled in the invoking channel.
type ModifyDisabledActionError struct {
	// Name is the name of the Action.
	Name string
}

var _ error = new(ModifyDisabledActionError)

func NewModifyDisabledActionError(actionName string) *ModifyDisabledActionError {
	return &ModifyDisabledActionError{Name: actionName}
}

func (e *ModifyDisabledActionError) Error() string {
	return fmt.Sprintf("action: the %s action is disabled and cannot be modified", e.Name)
}

func (e *ModifyDisabledActionError) As(target interface{}) bool {
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

func (e *ModifyDisabledActionError) AsUserError() *errors.UserError {
	return errors.NewUserErrorf("`%s` is disabled in this channel and, therefore, cannot be modified.", e.Name)
}

// =============================================================================
// AlreadyDisabledError
// =====================================================================================

// AlreadyDisabledError is the error returned if a user attempts to disable an
// Action that is already disabled.
//
// It makes itself available as a *errors.UserError when calling As.
type AlreadyDisabledError struct {
	// Name is the name of the Action.
	Name string
}

var _ error = new(AlreadyDisabledError)

func NewAlreadyDisabledError(actionName string) *AlreadyDisabledError {
	return &AlreadyDisabledError{Name: actionName}
}

func (e *AlreadyDisabledError) Error() string {
	return fmt.Sprintf("action: the %s action is already disabled in the invoking channel", e.Name)
}

func (e *AlreadyDisabledError) As(target interface{}) bool {
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

func (e *AlreadyDisabledError) AsUserError() *errors.UserError {
	return errors.NewUserErrorf("`%s` is already disabled in this channel.", e.Name)
}
