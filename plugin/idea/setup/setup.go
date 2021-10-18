// Package setup provides the setup command, used to set up a channel to use
// it to brainstorm ideas.
package setup

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/impl/arg"
	"github.com/mavolin/adam/pkg/impl/command"
	"github.com/mavolin/adam/pkg/impl/restriction"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/duration"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/internal/morearg"
	"github.com/mavolin/stormy/internal/stdcolor"
	"github.com/mavolin/stormy/internal/utils/deleteall"
	"github.com/mavolin/stormy/internal/utils/wizard"
)

type Setup struct {
	command.Meta
	bot.MiddlewareManager

	repo Repository
}

var _ plugin.Command = new(Setup)

func New(r Repository) *Setup {
	cmd := &Setup{
		Meta: command.Meta{
			Name:             "setup",
			ShortDescription: "Set up a channel for brainstorming.",
			LongDescription: "Set up a channel to brainstorm in.\n" +
				"If you're setting up your first channel, don't specify any flags and use the interactive setup. " +
				"If you're a pro already, quickly set up new channels by using flags for configuration. " +
				"Use `-use-defaults` to skip interactive setup and use the default settings, i.e. no flags.",
			Args: &arg.Config{
				Flags: []arg.Flag{
					{
						Name: "use-defaults",
						Type: arg.Switch,
						Description: "Skip the interactive setup and use the defaults for all options. " +
							"This flag cannot be used in conjunction with any other flag.",
					},
					{
						Name:    "vote-type",
						Aliases: []string{"type", "vt"},
						Type: arg.Choice{
							{Name: "thumbs", Value: Thumbs},
							{Name: "2 emojis", Value: TwoEmojis},
							{Name: "3 emojis", Value: ThreeEmojis},
							{Name: "5 emojis", Value: FiveEmojis},
						},
						Default: morearg.Undefined,
						Description: "The type of vote. Options are: `thumbs` (👍, 👎), " +
							"`2 emojis` (😀, ☹)️, `3 emojis` (😀, 😐, ☹), and " +
							"`5 emojis` (😀, 🙂, 😐, 🙁, ☹). " +
							"Defaults to `thumbs`.",
					},
					{
						Name:        "vote-duration",
						Aliases:     []string{"duration", "vd"},
						Type:        arg.SimpleDuration,
						Default:     morearg.Undefined,
						Description: "The duration to allow voting for. Defaults to infinity (`0`).",
					},
					{
						Name:    "anonymous",
						Aliases: []string{"a"},
						Type:    arg.Switch,
						Description: "Whether username and profile picture shall be omitted from the post. " +
							"Toggles anonymity when editing settings.",
					},
					{
						Name:        "color",
						Aliases:     []string{"colour", "c"},
						Type:        morearg.Color,
						Default:     morearg.Undefined,
						Description: "The color of the embed. Defaults to `44c5f3`.",
					},
					{
						Name:    "thumbnail",
						Aliases: []string{"t"},
						Type:    arg.SimpleLink,
						Description: "The thumbnail to use. Must be a link to an image." +
							"Defaults to none.",
					},
					{
						Name:        "rm-thumbnail",
						Type:        arg.Switch,
						Description: "Remove the current thumbnail. Can only be used when editing.",
					},
				},
			},
			ChannelTypes:   plugin.GuildTextChannels,
			BotPermissions: discord.PermissionSendMessages,
			Restrictions:   restriction.UserPermissions(discord.PermissionManageChannels),
		},
		repo: r,
	}

	cmd.AddMiddleware(deleteall.NewMiddleware(30 * time.Second))
	return cmd
}

func (setup *Setup) Invoke(s *state.State, ctx *plugin.Context) (interface{}, error) {
	settings, err := setup.repo.IdeaChannelSettings(ctx.ChannelID)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		return setup.firstTimeSetup(s, ctx)
	}

	return setup.modifySetup(s, ctx, ChannelSettings{})
}

func (setup *Setup) firstTimeSetup(s *state.State, ctx *plugin.Context) (interface{}, error) {
	var set ChannelSettings

	if len(ctx.RawArgs()) == 0 && !ctx.Flags.Bool("use-defaults") {
		w := wizard.New(s, ctx)

		w.AddStep(voteTypeStep(&set.VoteType))
		w.AddStep(voteDurationStep(&set.VoteDuration))
		w.AddStep(anonymousStep(&set.Anonymous))
		w.AddStep(colorStep(&set.Color))
		w.AddStep(thumbnailStep(&set.Thumbnail))

		if err := w.Start(); err != nil {
			return nil, err
		}
	} else {
		set = ChannelSettings{
			VoteType:     morearg.FlagOrDefault(ctx, "vote-type", Thumbs).(VoteType),
			VoteDuration: morearg.FlagOrDefault(ctx, "vote-duration", time.Duration(0)).(time.Duration),
			Anonymous:    ctx.Flags.Bool("anonymous"),
			Color:        morearg.FlagOrDefault(ctx, "color", stdcolor.Default).(discord.Color),
			Thumbnail:    ctx.Flags.String("thumbnail"),
		}
	}

	return msgbuilder.NewEmbed().
		WithTitle("All Set!").
		WithColor(stdcolor.Green).
		WithDescription("Every message sent in this channel will be turned into an idea users can vote on, " +
			"starting now.\n" +
			"The first line of each message will be turned into the title of the idea. " +
			"The remaining paragraphs are its description. Voting will stop after the voting duration has passed, " +
			"and the results will be added to the idea. If you choose to never stop voting, obviously, " +
			"no results will be posted. However, you can always have a look at the current voting."), nil
}

func (setup *Setup) modifySetup(
	s *state.State, ctx *plugin.Context, oldSettings ChannelSettings,
) (_ interface{}, err error) {
	if ctx.Flags.Bool("use-defaults") {
		return nil, errors.NewUserError("The `use-defaults` flag can only be used when setting up a new channel. " +
			"However, this channel is already set up.")
	}

	var newSettings ChannelSettings

	if len(ctx.RawArgs()) == 0 {
		newSettings, err = setup.modifySetupInteractive(s, ctx, oldSettings)
		if err != nil {
			return nil, err
		}
	} else {
		newSettings = setup.modifySetupFlags(ctx, oldSettings)
	}

	if oldSettings.Equals(newSettings) {
		return msgbuilder.NewEmbed().
			WithTitle("Nothing Changed").
			WithColor(stdcolor.Yellow).
			WithDescription("The changes you made are identical to the current settings."), nil
	}

	var ok bool

	_, err = msgbuilder.New(s, ctx).
		WithEmbed(setup.modifyChangeList(oldSettings, newSettings)).
		WithAwaitedComponent(msgbuilder.NewActionRow(&ok).
			With(msgbuilder.NewButton(discord.SuccessButton, "Yep, all good!", true)).
			With(msgbuilder.NewButton(discord.DangerButton, "I changed my mind.", false))).
		ReplyAndAwait(30 * time.Second)
	if err != nil {
		return nil, err
	}

	if !ok {
		return msgbuilder.NewEmbed().
			WithTitle("Changes discarded").
			WithColor(stdcolor.Yellow).
			WithDescription("I've discarded the changes you made. Everything remains as is."), nil
	}

	if err := setup.repo.SetIdeaChannelSettings(ctx.ChannelID, newSettings); err != nil {
		return nil, errors.WithDescription(err, "Something went wrong and I couldn't save your changes. "+
			"Try again in a bit.")
	}

	return msgbuilder.NewEmbed().
		WithTitle("Changes saved").
		WithColor(stdcolor.Green).
		WithDescription("I've save your changes."), nil
}

func (setup *Setup) modifyChangeList(oldS, newS ChannelSettings) *msgbuilder.EmbedBuilder {
	embed := msgbuilder.NewEmbed().
		WithTitle("Proceed?").
		WithColor(stdcolor.Default).
		WithDescription("You are about to change the following settings. Do you wish to proceed?")

	if oldS.VoteType != newS.VoteType {
		embed.WithInlinedField("Vote Type", fmt.Sprintf("`%s` ➡ `%s`", oldS.VoteType, newS.VoteType))
	}

	if oldS.VoteDuration != newS.VoteDuration {
		embed.WithInlinedField("Vote Duration",
			fmt.Sprintf("`%s` ➡ `%s`", duration.Format(oldS.VoteDuration), duration.Format(newS.VoteDuration)))
	}

	if oldS.Anonymous != newS.Anonymous {
		embed.WithInlinedField("Anonymous", fmt.Sprintf("`%t` ➡ `%t`", oldS.Anonymous, newS.Anonymous))
	}

	if oldS.Color != newS.Color {
		embed.WithInlinedField("Color", fmt.Sprintf("`%06x` ➡ `%06x`", oldS.Color, newS.Color))
	}

	if oldS.Thumbnail != newS.Thumbnail {
		embed.WithInlinedField("Thumbnail", "changed")
	}

	return embed
}

func (setup *Setup) modifySetupInteractive(
	s *state.State, ctx *plugin.Context, set ChannelSettings,
) (ChannelSettings, error) {
	var steps []wizard.Step

	_, err := msgbuilder.New(s, ctx).
		WithEmbed(msgbuilder.NewEmbed().
			WithTitle("What do you want to change?").
			WithColor(stdcolor.Default).
			WithDescription("Please select the settings you want to change in this channel.")).
		WithAwaitedComponent(msgbuilder.NewSelect(&steps).
			WithBounds(1, 5).
			With(msgbuilder.NewSelectOption("Vote Type", voteTypeStep(&set.VoteType))).
			With(msgbuilder.NewSelectOption("Vote Duration", voteDurationStep(&set.VoteDuration))).
			With(msgbuilder.NewSelectOption("Anonymity", anonymousStep(&set.Anonymous))).
			With(msgbuilder.NewSelectOption("Color", colorStep(&set.Color))).
			With(msgbuilder.NewSelectOption("Thumbnail", thumbnailStep(&set.Thumbnail)))).
		ReplyAndAwait(30 * time.Second)
	if err != nil {
		return ChannelSettings{}, err
	}

	w := wizard.New(s, ctx)

	for _, s := range steps {
		w.AddStep(s)
	}

	return set, w.Start()
}

func (setup *Setup) modifySetupFlags(ctx *plugin.Context, set ChannelSettings) ChannelSettings {
	if d := ctx.Flags["vote-duration"]; d != morearg.Undefined {
		set.VoteDuration = d.(time.Duration)
	}

	if t := ctx.Flags["vote-type"]; t != morearg.Undefined {
		set.VoteType = t.(VoteType)
	}

	if ctx.Flags.Bool("anonymous") {
		set.Anonymous = !set.Anonymous
	}

	if c := ctx.Flags["color"]; c != morearg.Undefined {
		set.Color = c.(discord.Color)
	}

	if ctx.Flags.Bool("rm-thumbnail") {
		set.Thumbnail = ""
	} else if t := ctx.Flags["thumbnail"]; t != morearg.Undefined {
		set.Thumbnail = t.(string)
	}

	return set
}
