package action

import (
	"reflect"
	"testing"

	"github.com/mavolin/adam/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// AlreadyEnabledError
// =====================================================================================

func TestEnabledError_As(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name   string
		Target interface{}
	}{
		{
			Name:   "*errors.Error",
			Target: new(errors.Error),
		},
		{
			Name:   "**errors.UserError",
			Target: new(*errors.UserError),
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			expect := NewAlreadyEnabledError("abc")

			require.Truef(t, errors.As(expect, c.Target), "errors.As returned false for %T", c.Target)

			asErr := reflect.ValueOf(c.Target).Elem().Interface()
			assert.Equal(t, expect.AsUserError(), asErr)
		})
	}
}

// =============================================================================
// AlreadyDisabledError
// =====================================================================================

func TestDisabledError_As(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name   string
		Target interface{}
	}{
		{
			Name:   "*errors.Error",
			Target: new(errors.Error),
		},
		{
			Name:   "**errors.UserError",
			Target: new(*errors.UserError),
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			expect := NewAlreadyDisabledError("abc")

			require.Truef(t, errors.As(expect, c.Target), "errors.As returned false for %T", c.Target)

			asErr := reflect.ValueOf(c.Target).Elem().Interface()
			assert.Equal(t, expect.AsUserError(), asErr)
		})
	}
}
