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
	// InstanceNames returns the names of the instances running in the channel
	// with the passed id, or nil if the command is disabled.
	//
	// As a special case, if an empty slice is returned, the action is
	// considered as enabled and as only using a single instance per channel by
	// design.
	InstanceNames(channelID discord.ChannelID) ([]string, error)

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
