package list

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/impl/command"
	"github.com/mavolin/adam/pkg/impl/restriction"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/internal/common"
	"github.com/mavolin/stormy/pkg/action"
)

type List struct {
	command.Meta

	actions []action.Action
}

var _ plugin.Command = new(List)

func New(actions ...action.Action) *List {
	return &List{
		Meta: command.Meta{
			Name:             "list",
			Aliases:          []string{"ls"},
			ShortDescription: "List all of the actions enabled in this channel.",
			ChannelTypes:     plugin.GuildTextChannels,
			BotPermissions:   discord.PermissionSendMessages,
			Restrictions:     restriction.UserPermissions(discord.PermissionManageChannels),
		},
		actions: actions,
	}
}

func (l *List) Invoke(_ *state.State, ctx *plugin.Context) (interface{}, error) {
	var errOccurred bool

	var listBuilder strings.Builder
	listBuilder.Grow(2048) // size of an embed description

	for _, a := range l.actions {
		instances, err := a.InstanceNames(ctx.ChannelID)
		if err != nil {
			errOccurred = true
			ctx.HandleErrorSilently(err)
		}

		if instances == nil {
			continue
		}

		if listBuilder.Len() > 0 {
			listBuilder.WriteRune('\n')
		}

		listBuilder.WriteString("• ")
		listBuilder.WriteString(a.Name())

		if a.IsSingleInstance() {
			continue
		}

		for _, name := range instances {
			listBuilder.WriteRune('\n')

			// use an ideographic space for indention, as Discord strips
			// regular whitespace
			listBuilder.WriteString("\u3000\u3000‣ ")
			listBuilder.WriteString(name)
		}
	}

	if listBuilder.Len() == 0 && errOccurred {
		return nil, errors.NewUserError("I couldn't find any actions that are enabled, however, " +
			"I'm experiencing some technical difficulties so I might be wrong. Try again in a bit.")
	}

	if listBuilder.Len() == 0 {
		return "No actions are enabled in this channel.", nil
	}

	if errOccurred {
		return msgbuilder.NewEmbed().
			WithTitle("Enabled Actions").
			WithDescription("The following actions are enabled in this channel."+
				"This list might be incomplete, as I'm currently experiencing some problems.").
			WithColor(common.EnabledColor).
			WithField("Actions", listBuilder.String()), nil
	}

	return msgbuilder.NewEmbed().
		WithTitle("Enabled Actions").
		WithDescription("The following actions are enabled in this channel.").
		WithColor(common.EnabledColor).
		WithField("Actions", listBuilder.String()), nil
}
