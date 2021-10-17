package list

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/impl/command"
	"github.com/mavolin/adam/pkg/impl/restriction"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/internal/stdcolor"
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
			ShortDescription: "List all enabled and all available actions in this channel.",
			ChannelTypes:     plugin.GuildTextChannels | plugin.GuildNewsChannels,
			BotPermissions:   discord.PermissionSendMessages,
			Restrictions:     restriction.UserPermissions(discord.PermissionManageChannels),
		},
		actions: actions,
	}
}

func (l *List) Invoke(_ *state.State, ctx *plugin.Context) (interface{}, error) {
	var enabledBuilder, disabledBuilder strings.Builder
	enabledBuilder.Grow(2048) // size of an embed field
	disabledBuilder.Grow(2048)

	var errOccurred bool

	for _, a := range l.actions {
		instances, err := a.GetInstanceNames(ctx.ChannelID)
		if err != nil {
			ctx.HandleErrorSilently(err)
			errOccurred = true
			continue
		}

		if instances == nil {
			writeListItem(&disabledBuilder, a.GetName())
			continue
		}

		writeListItem(&enabledBuilder, a.GetName())

		if a.IsSingleInstance() {
			continue
		}

		for _, name := range instances {
			enabledBuilder.WriteRune('\n')

			// use an ideographic space for indention, as Discord strips
			// regular whitespace
			enabledBuilder.WriteString("\u3000\u3000‣ ")
			enabledBuilder.WriteString(name)
		}
	}

	desc := "Below is a list of all enabled and all disabled actions in this channel."
	if errOccurred {
		desc += " However, this list is incomplete, as I'm currently experiencing some technical difficulties."
	}

	resp := msgbuilder.NewEmbed().
		WithTitle("Actions").
		WithDescription(desc).
		WithColor(stdcolor.Default)

	if enabledBuilder.Len() > 0 {
		resp.WithField("Enabled Actions", enabledBuilder.String())
	}

	if disabledBuilder.Len() > 0 {
		resp.WithField("Disabled Actions", disabledBuilder.String())
	}

	return resp, nil
}

func writeListItem(b *strings.Builder, item string) {
	if b.Len() > 0 {
		b.WriteRune('\n')
	}

	b.WriteString("• ")
	b.WriteString(item)
}
