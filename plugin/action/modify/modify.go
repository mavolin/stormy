package modify

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

type Modify struct {
	command.Meta

	actions []action.Action
}

var _ plugin.Command = new(Modify)

func New(actions ...action.Action) *Modify {
	return &Modify{
		Meta: command.Meta{
			Name:             "modify",
			ShortDescription: "Modify an action in the calling channel.",
			ChannelTypes:     plugin.GuildTextChannels,
			BotPermissions:   discord.PermissionSendMessages,
			Restrictions:     restriction.UserPermissions(discord.PermissionManageChannels),
		},
		actions: actions,
	}
}

func (e *Modify) Invoke(s *state.State, ctx *plugin.Context) (interface{}, error) {
	var selectedAction action.Action
	actionSelect := msgbuilder.NewSelect(&selectedAction)

	for _, selectableAction := range e.actions {
		actionSelect.With(msgbuilder.NewSelectOption(selectableAction.GetName(), selectableAction))
	}

	_, err := msgbuilder.New(s, ctx).
		WithContent("Please select the action you would like to modify.").
		WithAwaitedComponent(actionSelect).
		ReplyAndAwait(15 * time.Second)
	if err != nil {
		return nil, err
	}

	c, err := ctx.Channel()
	if err != nil {
		return nil, err
	}

	if !selectedAction.GetChannelTypes().Has(c.Type) {
		return nil, plugin.NewChannelTypeError(selectedAction.GetChannelTypes())
	}

	return nil, selectedAction.Modify(ctx)
}
