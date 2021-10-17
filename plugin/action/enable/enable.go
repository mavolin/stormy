// Package enable provides the enable command.
package enable

import (
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/impl/command"
	"github.com/mavolin/adam/pkg/impl/restriction"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/pkg/action"
)

type Enable struct {
	command.Meta

	actions []action.Action
}

var _ plugin.Command = new(Enable)

func New(actions ...action.Action) *Enable {
	return &Enable{
		Meta: command.Meta{
			Name:             "enable",
			Aliases:          []string{"off"},
			ShortDescription: "Enable an action in the calling channel.",
			ChannelTypes:     plugin.GuildTextChannels | plugin.GuildNewsChannels,
			BotPermissions:   discord.PermissionSendMessages,
			Restrictions:     restriction.UserPermissions(discord.PermissionManageChannels),
		},
		actions: actions,
	}
}

func (e *Enable) Invoke(s *state.State, ctx *plugin.Context) (interface{}, error) {
	selectedAction, err := e.selectAction(s, ctx)
	if err != nil {
		return nil, err
	}

	if err = e.ensureChannelTypes(ctx, selectedAction); err != nil {
		return nil, err
	}

	return nil, selectedAction.Enable(ctx)
}

func (e *Enable) selectAction(s *state.State, ctx *plugin.Context) (action.Action, error) {
	var selectedAction action.Action
	actionSelect := msgbuilder.NewSelect(&selectedAction)

	for _, selectableAction := range e.actions {
		actionSelect.With(msgbuilder.NewSelectOption(selectableAction.GetName(), selectableAction))
	}

	_, err := msgbuilder.New(s, ctx).
		WithContent("Please select the action you would like to enable.").
		WithAwaitedComponent(actionSelect).
		ReplyAndAwait(15 * time.Second)
	return selectedAction, err
}

func (e *Enable) ensureChannelTypes(ctx *plugin.Context, selectedAction action.Action) error {
	c, err := ctx.Channel()
	if err != nil {
		return err
	}

	if !selectedAction.GetChannelTypes().Has(c.Type) {
		return plugin.NewChannelTypeError(selectedAction.GetChannelTypes())
	}

	return nil
}
