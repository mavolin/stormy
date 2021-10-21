// Package how provides the how command, explaining how to post ideas.
package how

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/impl/command"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/internal/stdcolor"
)

type How struct {
	command.Meta
}

func New() *How {
	return &How{
		Meta: command.Meta{
			Name:             "how",
			Aliases:          []string{"explain", "help"},
			ShortDescription: "Get an explanation of how posting ideas works.",
			Hidden:           false,
			ChannelTypes:     plugin.GuildTextChannels,
			BotPermissions:   discord.PermissionSendMessages,
		},
	}
}

var _ plugin.Command = new(How)

func (h *How) Invoke(*state.State, *plugin.Context) (interface{}, error) {
	return msgbuilder.NewEmbed().
			WithTitle("Posting Ideas").
			WithColor(stdcolor.Default).
			WithDescription("Posting an idea is easy. "+
				"The first line of each message will be turned into the title of the idea. "+
				"The remaining paragraphs are its description. Voting will stop after the voting duration has passed, "+
				"and the results will be added to the idea post.").
			WithField("Sections", "You can also add up to 12 sections to your idea. "+
				"Each section can be upvoted to signal that this part of idea is wanted.\n"+
				"To create a section, start a new line with a `#`. "+
				"You can also create mutually exclusive groups of sections by amending a number to the `#`, e.g. `#1`. "+
				"Within a group, users can only vote for one section."+
				"If a user votes for a second section within a group, the previous vote is removed.\n"+
				"Groups have no special markings when posted."),
		nil
}
