// Package action provides the abstraction of an action.
package action

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/state"
)

// Action is the abstraction of an action, as used by the action plugin.
type Action interface {
	// GetName returns the name of the action.
	GetName() string

	// IsSingleInstance reports whether the Action is designed to only run a
	// single instance per channel.
	IsSingleInstance() bool
	// GetInstanceNames returns the names of the instances running in the
	// channel with the passed id, or nil if the command is disabled.
	GetInstanceNames(channelID discord.ChannelID) ([]string, error)

	// GetChannelTypes returns the channel types the action works in.
	// The ChannelTypes must be either plugin.GuildTextChannels,
	// plugin.GuildNewsChannels, or both.
	GetChannelTypes() plugin.ChannelTypes

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

// ActionMeta is an embeddable struct that can be used to provide the metadata
// getters of an Action.
type ActionMeta struct {
	Name           string
	SingleInstance bool
	ChannelTypes   plugin.ChannelTypes
}

func (m ActionMeta) GetName() string                      { return m.Name }
func (m ActionMeta) IsSingleInstance() bool               { return m.SingleInstance }
func (m ActionMeta) GetChannelTypes() plugin.ChannelTypes { return m.ChannelTypes }

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
