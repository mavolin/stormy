package setup

import (
	"net/url"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/duration"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/internal/stdcolor"
	"github.com/mavolin/stormy/pkg/utils/wizard"
)

func voteTypeStep(typ *VoteType) wizard.Step {
	return wizard.Step{
		Question: func(s *state.State, ctx *plugin.Context) *msgbuilder.Builder {
			return msgbuilder.New(s, ctx).
				WithEmbed(msgbuilder.NewEmbed().
					WithTitle(wizard.Title(ctx, "Type of Vote")).
					WithColor(stdcolor.Default).
					WithDescription("What type of scale should be used to vote?")).
				WithComponent(msgbuilder.NewSelect(typ).
					WithDefault(msgbuilder.NewSelectOption("👍, and 👎", Thumbs)).
					With(msgbuilder.NewSelectOption("😀 and ☹", TwoEmojis)).
					With(msgbuilder.NewSelectOption("😀, 😐, and ☹", ThreeEmojis)).
					With(msgbuilder.NewSelectOption("😀, 🙂, 😐, 🙁, and ☹", FiveEmojis))).
				WithAwaitedComponent(msgbuilder.NewActionRow(new(struct{})).
					With(msgbuilder.NewButton(discord.SuccessButton, "Done", struct{}{})))
		},
		WaitFor: 25 * time.Second,
	}
}

func voteDurationStep(voteDuration *time.Duration) wizard.Step {
	var voteDurationMsg discord.Message
	var forever bool

	return wizard.Step{
		Question: func(s *state.State, ctx *plugin.Context) *msgbuilder.Builder {
			return msgbuilder.New(s, ctx).
				WithEmbed(msgbuilder.NewEmbed().
					WithTitle(wizard.Title(ctx, "Vote Duration")).
					WithColor(stdcolor.Default).
					WithDescription("How long should users be able to vote on an idea?\n"+
						"Type a duration like `24h` or `1d 12h` or click the `Forever` button. "+
						"The duration must be at least one minute, and may be longer than a week.")).
				WithAwaitedResponse(&voteDurationMsg, 20*time.Second, 10*time.Second).
				WithAwaitedComponent(msgbuilder.NewActionRow(&forever).
					With(msgbuilder.NewButton(discord.PrimaryButton, "Forever", true).
						WithEmoji(discord.ButtonEmoji{Name: "♾"})))
		},
		WaitFor: 30 * time.Second,
		Validator: func(s *state.State, ctx *plugin.Context) (err error) {
			if forever {
				*voteDuration = 0
				return nil
			}

			*voteDuration, err = duration.Parse(voteDurationMsg.Content)
			if err != nil || *voteDuration <= 0 {
				return wizard.NewRetryErrorf("`%s` is not a valid duration.", voteDurationMsg.Content)
			}

			if *voteDuration < duration.Minute || *voteDuration > duration.Week {
				return wizard.NewRetryError("The duration must be between one minute and one week.")
			}

			return nil
		},
	}
}

func anonymousStep(anonymous *bool) wizard.Step {
	return wizard.Step{
		Question: func(s *state.State, ctx *plugin.Context) *msgbuilder.Builder {
			return msgbuilder.New(s, ctx).
				WithEmbed(msgbuilder.NewEmbed().
					WithTitle(wizard.Title(ctx, "Anonymous")).
					WithColor(stdcolor.Default).
					WithDescription("Shall the author of the idea remain anonymous?\n" +
						"If not their username and profile picture at that time will be included in the post.")).
				WithAwaitedComponent(msgbuilder.NewActionRow(anonymous).
					With(msgbuilder.NewButton(discord.PrimaryButton, "Anonymous", true).
						WithEmoji(discord.ButtonEmoji{Name: "🕵"})).
					With(msgbuilder.NewButton(discord.PrimaryButton, "Public", false).
						WithEmoji(discord.ButtonEmoji{Name: "🌐"})))
		},
		WaitFor: 20 * time.Second,
	}
}

func colorStep(color *discord.Color) wizard.Step {
	var colorMsg discord.Message

	return wizard.Step{
		Question: func(s *state.State, ctx *plugin.Context) *msgbuilder.Builder {
			return msgbuilder.New(s, ctx).
				WithEmbed(msgbuilder.NewEmbed().
					WithTitle(wizard.Title(ctx, "Color")).
					WithColor(stdcolor.Default).
					WithDescription("What color (the line at the left) shall the idea have, when posted?\n"+
						"Respond with a hexadecimal color such as `44c5f3`.")).
				WithAwaitedResponse(&colorMsg, 90*time.Second, 20*time.Second).
				WithAwaitedComponent(msgbuilder.NewActionRow(color).
					With(msgbuilder.NewButton(discord.PrimaryButton, "I don't care, use the default.",
						stdcolor.Default)))
		},
		WaitFor: 120 * time.Second,
		Validator: func(s *state.State, ctx *plugin.Context) error {
			if *color == stdcolor.Default {
				return nil
			}

			cInt, err := strconv.ParseInt(colorMsg.Content, 16, 32)
			if err != nil || len(colorMsg.Content) != 6 || cInt < 0x000000 || cInt > 0xffffff {
				return wizard.NewRetryErrorf("`%s` is not a valid color.", colorMsg.Content)
			}

			*color = discord.Color(cInt)
			return nil
		},
	}
}

func thumbnailStep(thumbnailURL *discord.URL) wizard.Step {
	var thumbnailMsg discord.Message
	var none bool

	return wizard.Step{
		Question: func(s *state.State, ctx *plugin.Context) *msgbuilder.Builder {
			return msgbuilder.New(s, ctx).
				WithEmbed(msgbuilder.NewEmbed().
					WithTitle(wizard.Title(ctx, "Thumbnail")).
					WithColor(stdcolor.Default).
					WithDescription("What thumbnail (little picture on the top right) should each post contain?\n"+
						"Please upload an image, send a link, or click `None`.")).
				WithAwaitedResponse(&thumbnailMsg, 120*time.Second, 20*time.Second).
				WithAwaitedComponent(msgbuilder.NewActionRow(&none).
					With(msgbuilder.NewButton(discord.PrimaryButton, "None", true)))
		},
		WaitFor: 180 * time.Second,
		Validator: func(s *state.State, ctx *plugin.Context) error {
			if none {
				return nil
			}

			if len(thumbnailMsg.Attachments) > 0 {
				*thumbnailURL = thumbnailMsg.Attachments[0].URL
				return nil
			}

			_, err := url.ParseRequestURI(thumbnailMsg.Content)
			if err != nil {
				return wizard.NewRetryError("The link you gave me is invalid.")
			}

			*thumbnailURL = thumbnailMsg.Content
			return nil
		},
	}
}
