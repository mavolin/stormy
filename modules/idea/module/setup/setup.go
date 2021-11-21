// Package setup provides the setup command, used to set up a channel to use
// it to brainstorm ideas.
package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/impl/arg"
	"github.com/mavolin/adam/pkg/impl/command"
	"github.com/mavolin/adam/pkg/impl/restriction"
	"github.com/mavolin/adam/pkg/impl/throttler"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/duration"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/internal/stdcolor"
	"github.com/mavolin/stormy/modules/idea/repository"
	"github.com/mavolin/stormy/pkg/morearg"
	"github.com/mavolin/stormy/pkg/utils/deleteall"
	"github.com/mavolin/stormy/pkg/utils/wizard"
)

type Setup struct {
	command.Meta
	bot.MiddlewareManager

	repo repository.ChannelSettingsRepository
}

var _ plugin.Command = new(Setup)

func New(r repository.ChannelSettingsRepository) *Setup {
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
							{Name: "thumbs", Value: repository.Thumbs},
							{Name: "2 emojis", Value: repository.TwoEmojis},
							{Name: "3 emojis", Value: repository.ThreeEmojis},
							{Name: "5 emojis", Value: repository.FiveEmojis},
						},
						Default: morearg.Undefined,
						Description: "The type of vote. Options are: `thumbs` (👍, 👎), " +
							"`2 emojis` (😀, ☹)️, `3 emojis` (😀, 😐, ☹), and " +
							"`5 emojis` (😀, 🙂, 😐, 🙁, ☹). " +
							"Defaults to `thumbs`.",
					},
					{
						Name:    "vote-duration",
						Aliases: []string{"duration", "vd"},
						Type: arg.Duration{
							Min: duration.Minute,
							Max: duration.Week,
						},
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
				},
			},
			ChannelTypes:   plugin.GuildTextChannels,
			BotPermissions: discord.PermissionSendMessages,
			Restrictions:   restriction.UserPermissions(discord.PermissionManageChannels),
			Throttler:      throttler.PerChannel(1, 5*time.Second),
		},
		repo: r,
	}

	cmd.AddMiddleware(deleteall.NewMiddleware(30 * time.Second))
	return cmd
}

func (setup *Setup) Invoke(s *state.State, pctx *plugin.Context) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	settings, err := setup.repo.IdeaChannelSettings(ctx, pctx.ChannelID)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		return setup.firstTimeSetup(s, pctx)
	}

	return setup.modifySetup(s, pctx, *settings)
}

func (setup *Setup) firstTimeSetup(s *state.State, pctx *plugin.Context) (interface{}, error) {
	var set repository.ChannelSettings

	if len(pctx.RawArgs()) == 0 && !pctx.Flags.Bool("use-defaults") {
		w := wizard.New(s, pctx)

		w.AddStep(voteTypeStep(&set.VoteType))
		w.AddStep(voteDurationStep(&set.VoteDuration))
		w.AddStep(anonymousStep(&set.Anonymous))
		w.AddStep(colorStep(&set.Color))

		if err := w.Start(); err != nil {
			if errors.Is(err, errors.Abort) {
				return msgbuilder.NewEmbed().
					WithTitle("Cancel").
					WithColor(stdcolor.Yellow).
					WithDescription("The setup has been cancelled."), nil
			}

			return nil, err
		}
	} else {
		set = repository.ChannelSettings{
			VoteType:     morearg.FlagOrDefault(pctx, "vote-type", repository.Thumbs).(repository.VoteType),
			VoteDuration: morearg.FlagOrDefault(pctx, "vote-duration", time.Duration(0)).(time.Duration),
			Anonymous:    pctx.Flags.Bool("anonymous"),
			Color:        morearg.FlagOrDefault(pctx, "color", stdcolor.Default).(discord.Color),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := setup.repo.SetIdeaChannelSettings(ctx, pctx.ChannelID, set)
	if err != nil {
		return nil, errors.WithDescription(err, "Something went wrong and I couldn't save the settings you choose. "+
			"Try again in a bit.")
	}

	return msgbuilder.NewEmbed().
		WithTitle("All Set!").
		WithColor(stdcolor.Green).
		WithDescription("Every message sent in this channel will be turned into an idea users can vote on, " +
			"starting now. If you need help with posting or formatting an idea, use `idea how`.\n" +
			"I'll automatically delete all setup related messages in 30 seconds."), nil
}

func (setup *Setup) modifySetup(
	s *state.State, pctx *plugin.Context, oldSettings repository.ChannelSettings,
) (_ interface{}, err error) {
	if pctx.Flags.Bool("use-defaults") {
		return nil, errors.NewUserError("The `use-defaults` flag can only be used when setting up a new channel. " +
			"However, this channel is already set up.")
	}

	var newSettings repository.ChannelSettings

	if len(pctx.RawArgs()) == 0 {
		newSettings, err = setup.modifySetupInteractive(s, pctx, oldSettings)
		if err != nil {
			return nil, err
		}
	} else {
		newSettings = setup.modifySetupFlags(pctx, oldSettings)
	}

	if oldSettings == newSettings {
		return msgbuilder.NewEmbed().
			WithTitle("Nothing Changed").
			WithColor(stdcolor.Yellow).
			WithDescription("The changes you made are identical to the current settings."), nil
	}

	var ok bool

	_, err = msgbuilder.New(s, pctx).
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := setup.repo.SetIdeaChannelSettings(ctx, pctx.ChannelID, newSettings); err != nil {
		return nil, errors.WithDescription(err, "Something went wrong and I couldn't save your changes. "+
			"Try again in a bit.")
	}

	return msgbuilder.NewEmbed().
		WithTitle("Changes saved").
		WithColor(stdcolor.Green).
		WithDescription("I've saved your changes."), nil
}

func (setup *Setup) modifySetupInteractive(
	s *state.State, pctx *plugin.Context, set repository.ChannelSettings,
) (repository.ChannelSettings, error) {
	var steps []wizard.Step
	var done bool

	_, err := msgbuilder.New(s, pctx).
		WithEmbed(msgbuilder.NewEmbed().
			WithTitle("What do you want to change?").
			WithColor(stdcolor.Default).
			WithDescription("Please select the settings you want to change in this channel.")).
		WithComponent(msgbuilder.NewSelect(&steps).
			WithBounds(1, 5).
			With(msgbuilder.NewSelectOption("Vote Type", voteTypeStep(&set.VoteType))).
			With(msgbuilder.NewSelectOption("Vote Duration", voteDurationStep(&set.VoteDuration))).
			With(msgbuilder.NewSelectOption("Anonymity", anonymousStep(&set.Anonymous))).
			With(msgbuilder.NewSelectOption("Color", colorStep(&set.Color)))).
		WithAwaitedComponent(msgbuilder.NewActionRow(&done).
			With(msgbuilder.NewButton(discord.SuccessButton, "Done", true)).
			With(msgbuilder.NewButton(discord.DangerButton, "Cancel", false))).
		ReplyAndAwait(30 * time.Second)
	if err != nil {
		return repository.ChannelSettings{}, err
	}

	if !done {
		return repository.ChannelSettings{}, errors.Abort
	}

	w := wizard.New(s, pctx)
	for _, s := range steps {
		w.AddStep(s)
	}

	return set, w.Start()
}

func (setup *Setup) modifySetupFlags(ctx *plugin.Context, set repository.ChannelSettings) repository.ChannelSettings {
	if d := ctx.Flags["vote-duration"]; d != morearg.Undefined {
		set.VoteDuration = d.(time.Duration)
	}

	if t := ctx.Flags["vote-type"]; t != morearg.Undefined {
		set.VoteType = t.(repository.VoteType)
	}

	if ctx.Flags.Bool("anonymous") {
		set.Anonymous = !set.Anonymous
	}

	if c := ctx.Flags["color"]; c != morearg.Undefined {
		set.Color = c.(discord.Color)
	}

	return set
}

func (setup *Setup) modifyChangeList(oldS, newS repository.ChannelSettings) *msgbuilder.EmbedBuilder {
	embed := msgbuilder.NewEmbed().
		WithTitle("Everything in Order?").
		WithColor(stdcolor.Default).
		WithDescription("You are about to change the following settings. do you wish to proceed?")

	if oldS.VoteType != newS.VoteType {
		embed.WithInlinedField("Vote Type", fmt.Sprintf("`%s` ➡ `%s`", oldS.VoteType, newS.VoteType))
	}

	if oldS.VoteDuration != newS.VoteDuration {
		oldDuration := "forever"
		if oldS.VoteDuration > 0 {
			oldDuration = duration.Format(oldS.VoteDuration)
		}

		newDuration := "forever"
		if newS.VoteDuration > 0 {
			newDuration = duration.Format(newS.VoteDuration)
		}

		embed.WithInlinedField("Vote Duration", fmt.Sprintf("`%s` ➡ `%s`", oldDuration, newDuration))
	}

	if oldS.Anonymous != newS.Anonymous {
		oldAnonymous := "anonymous"
		if !oldS.Anonymous {
			oldAnonymous = "public"
		}

		newAnonymous := "anonymous"
		if !newS.Anonymous {
			newAnonymous = "public"
		}

		embed.WithInlinedField("Anonymous", fmt.Sprintf("`%s` ➡ `%s`", oldAnonymous, newAnonymous))
	}

	if oldS.Color != newS.Color {
		embed.WithInlinedField("Color", fmt.Sprintf("`%06x` ➡ `%06x`", oldS.Color, newS.Color))
	}

	return embed
}
