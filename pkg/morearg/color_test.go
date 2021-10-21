package morearg

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColor_Parse(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var expect discord.Color = 0x44c5f3

		ctx := &plugin.ParseContext{Raw: "44c5f3"}

		actual, err := Color.Parse(nil, ctx)
		require.NoError(t, err)
		assert.Equal(t, expect, actual)
	})

	failureCases := []struct {
		Name string
		Raw  string
	}{
		{Name: "too short", Raw: "fff"},
		{Name: "too long", Raw: "ff00ff1"},
		{Name: "invalid characters", Raw: "ffggff"},
	}

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		for _, c := range failureCases {
			t.Run(c.Name, func(t *testing.T) {
				t.Parallel()

				ctx := &plugin.ParseContext{Raw: c.Raw}

				_, err := Color.Parse(nil, ctx)
				assert.ErrorAs(t, err, new(*plugin.ArgumentError))
			})
		}
	})
}
