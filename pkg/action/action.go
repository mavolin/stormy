// Package action provides the abstraction of an action.
package action

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/state"
)

// Action is the abstraction of an action, as used by the action plugin.
type Action interface {
	// Name returns the name of the action.
	Name() string

	// IsSingleInstance reports whether the Action is designed to only run a
	// single instance per channel.
	IsSingleInstance() bool
	// InstanceNames returns the names of the instances running in the channel
	// with the passed id, or nil if the command is disabled.
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

// OnceCommander is an interface that can optionally be implemented by actions
// if they support running the action once.
type OnceCommander interface {
	// OnceCommand returns the OnceCommand used to perform an action a single
	// time.
	OnceCommand() OnceCommand
}

type OnceCommand struct {
	ShortDescription string
	LongDescription  string
	Args             plugin.ArgConfig
	ArgParser        plugin.ArgParser
	ExampleArgs      plugin.ExampleArgs
	ChannelTypes     plugin.ChannelTypes
	BotPermissions   discord.Permissions
	Restrictions     plugin.RestrictionFunc
	Throttler        plugin.Throttler

	Invoke func(*state.State, *plugin.Context) (interface{}, error)
}
