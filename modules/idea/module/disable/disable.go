package disable

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/impl/command"
	"github.com/mavolin/adam/pkg/impl/restriction"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/internal/stdcolor"
	"github.com/mavolin/stormy/modules/idea/repository"
)

type Disable struct {
	command.Meta

	repo repository.ChannelSettingsRepository
}

var _ plugin.Command = new(Disable)

func New(r repository.ChannelSettingsRepository) *Disable {
	return &Disable{
		Meta: command.Meta{
			Name:             "disable",
			ShortDescription: "Disables brainstorming in this channel.",
			ChannelTypes:     plugin.GuildTextChannels,
			BotPermissions:   discord.PermissionSendMessages,
			Restrictions:     restriction.UserPermissions(discord.PermissionManageChannels),
		},
		repo: r,
	}
}

func (d *Disable) Invoke(_ *state.State, ctx *plugin.Context) (interface{}, error) {
	if err := d.repo.DisableIdeaChannel(ctx.ChannelID); err != nil {
		return nil, err
	}

	return msgbuilder.NewEmbed().
		WithTitle("Brainstorming").
		WithColor(stdcolor.Green).
		WithDescription("Brainstorming has been successfully disabled in this channel."), nil
}
