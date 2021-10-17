// Package do provides the do module that allows performing OnceCommands of
// Actions that support it.
package do

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/i18n"
	"github.com/mavolin/adam/pkg/impl/module"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/pkg/action"
)

func New(actions ...action.Action) *module.Module {
	mod := module.New(module.Meta{
		Name:             "do",
		ShortDescription: "Perform an action once.",
		LongDescription: "Some actions support running them once, without having to enable them." +
			" This module provides commands to run all of those actions.",
	})

	for _, a := range actions {
		if oc, ok := a.(action.OnceCommander); ok {
			mod.AddCommand(&onceCommand{name: a.GetName(), OnceCommand: oc.OnceCommand()})
		}
	}

	return mod
}

type onceCommand struct {
	name string
	action.OnceCommand
}

func (o *onceCommand) GetName() string                                   { return o.name }
func (o *onceCommand) GetAliases() []string                              { return nil }
func (o *onceCommand) GetShortDescription(*i18n.Localizer) string        { return o.ShortDescription }
func (o *onceCommand) GetLongDescription(*i18n.Localizer) string         { return o.LongDescription }
func (o *onceCommand) GetArgs() plugin.ArgConfig                         { return o.Args }
func (o *onceCommand) GetArgParser() plugin.ArgParser                    { return o.ArgParser }
func (o *onceCommand) GetExampleArgs(*i18n.Localizer) plugin.ExampleArgs { return o.ExampleArgs }
func (o *onceCommand) IsHidden() bool                                    { return false }
func (o *onceCommand) GetChannelTypes() plugin.ChannelTypes              { return o.ChannelTypes }
func (o *onceCommand) GetBotPermissions() discord.Permissions            { return o.BotPermissions }

func (o *onceCommand) IsRestricted(s *state.State, ctx *plugin.Context) error {
	return o.Restrictions(s, ctx)
}

func (o *onceCommand) GetThrottler() plugin.Throttler { return o.Throttler }

func (o *onceCommand) Invoke(s *state.State, ctx *plugin.Context) (interface{}, error) {
	return o.OnceCommand.Invoke(s, ctx)
}
