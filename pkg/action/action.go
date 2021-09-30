// Package action provides the abstraction of an action.
package action

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/plugin"
)

// Action is the abstraction of an action, as used by the action plugin.
type Action interface {
	// Name returns the name of the action.
	Name() string
	// IsEnabled reports whether the action is set up in the channel with the
	// passed id.
	IsEnabled(channelID discord.ChannelID) bool

	// Enable enables the action in the invoking channel.
	//
	// If the action is already enabled, a *AlreadyEnabledError is returned.
	Enable(ctx *plugin.Context) error
	// Modify modifies the settings of an action in the invoking channel.
	//
	// If the action is disabled, a *ModifyDisabledActionError is returned.
	Modify(ctx *plugin.Context) error
	// Disable disables the action in the invoking channel.
	//
	// If the action is already disabled, a *AlreadyDisabledError is returned.
	Disable(ctx *plugin.Context) error
}
