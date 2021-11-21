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
			"The remaining paragraphs are its optional description. "+
			"The description ends with the first section or the end of the message. "+
			"An idea needs either a description or at least one section.\n"+
			"Voting will stop after the voting duration has passed, "+
			"and the results will be added to the idea post.").
		WithField("Sections", "You can also add up to 15 sections to your idea. "+
			"Each section can be upvoted to signal that this part of idea is wanted.\n"+
			"To create a section, start a new line with a `*`. ").
		WithField("Groups",
			"You can also create groups of sections. "+
				"Within a group sections are considered mutually exclusive, "+
				"meaning users can only vote for a single section within that group.\n"+
				"You can start a group by writing `# My Group Title` on a new line. "+
				"All sections that were written before the first group won't be mutually exclusive. "+
				"To start a second group, simply write a new title\n"+
				"**Example:**\n"+
				"```\n"+
				"My Title\n\n"+
				"* My first regular section.\n\n"+
				"# My Group\n"+
				"* My first group section.\n"+
				"* My second group section.\n\n"+
				"# My Other Group\n"+
				"* My other group section."+
				"```\n"), nil
}
