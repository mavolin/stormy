package format

import (
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mavolin/stormy/internal/errhandler"
)

func TestParser_Parse(t *testing.T) {
	t.Parallel()

	successCases := []struct {
		name   string
		in     string
		expect parserIdea
	}{
		{
			name: "title and description",
			in: "My Title\n" +
				"My description.",
			expect: parserIdea{
				title: "My Title",
				desc:  "My description.",
			},
		},
		{
			name: "256 char title",
			in:   strings.Repeat("a", 256) + "\n.",
			expect: parserIdea{
				title: strings.Repeat("a", 256),
				desc:  ".",
			},
		},
		{
			name: "multiples newlines after title",
			in: "My Title\n\n\n" +
				"My description.",
			expect: parserIdea{
				title: "My Title",
				desc:  "My description.",
			},
		},
		{
			name: "title and sections",
			in: "My title\n" +
				"* My first section.\n" +
				"* My second section.",
			expect: parserIdea{
				title:    "My title",
				sections: []string{"My first section.", "My second section."},
			},
		},
		{
			name: "no space after asterisk",
			in: "My Title\n" +
				"*Abc",
			expect: parserIdea{
				title:    "My Title",
				sections: []string{"Abc"},
			},
		},
		{
			name: "1024 char section",
			in: ".\n" +
				"* " + strings.Repeat("a", 1024),
			expect: parserIdea{
				title:    ".",
				sections: []string{strings.Repeat("a", 1024)},
			},
		},
		{
			name: "multiple newlines after section",
			in: "My title\n" +
				"* My first section.\n\n" +
				"* My second section.",
			expect: parserIdea{
				title:    "My title",
				sections: []string{"My first section.", "My second section."},
			},
		},
		{
			name: "title, description and sections",
			in: "My title\n" +
				"My description.\n" +
				"* My first section.\n" +
				"* My second section.",
			expect: parserIdea{
				title:    "My title",
				desc:     "My description.",
				sections: []string{"My first section.", "My second section."},
			},
		},
		{
			name: "group",
			in: "title\n" +
				"# My Group Title\n" +
				"* My section content.",
			expect: parserIdea{
				title: "title",
				groups: []group{
					{
						title:    "My Group Title",
						sections: []string{"My section content."},
					},
				},
			},
		},
		{
			name: "no space after hash",
			in: "My Title\n" +
				"#Abc\n" +
				"* Def",
			expect: parserIdea{
				title: "My Title",
				groups: []group{
					{title: "Abc", sections: []string{"Def"}},
				},
			},
		},
		{
			name: "256 char section title",
			in:   ".\n# " + strings.Repeat("a", 256) + "\n* .",
			expect: parserIdea{
				title: ".",
				groups: []group{
					{
						title:    strings.Repeat("a", 256),
						sections: []string{"."},
					},
				},
			},
		},
		{
			name: "multiple groups",
			in: "title\n" +
				"# My Section Title\n" +
				"* My first section.\n" +
				"* My second section.\n" +
				"# My Other Section Title\n" +
				"* My last section.",
			expect: parserIdea{
				title: "title",
				groups: []group{
					{
						title:    "My Section Title",
						sections: []string{"My first section.", "My second section."},
					},
					{
						title:    "My Other Section Title",
						sections: []string{"My last section."},
					},
				},
			},
		},
		{
			name: "everything",
			in: "My Title\n" +
				"My description.\n" +
				"* My first global section.\n" +
				"* My second global section.\n" +
				"# My First Group\n" +
				"* My first group section.\n" +
				"* My second group section.\n" +
				"# My Second Group\n" +
				"* My first other group section.\n" +
				"* My second other group section.",
			expect: parserIdea{
				title:    "My Title",
				desc:     "My description.",
				sections: []string{"My first global section.", "My second global section."},
				groups: []group{
					{
						title:    "My First Group",
						sections: []string{"My first group section.", "My second group section."},
					},
					{
						title: "My Second Group",
						sections: []string{
							"My first other group section.",
							"My second other group section.",
						},
					},
				},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		for _, c := range successCases {
			c := c
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				p := newParser(&event.MessageCreate{
					MessageCreateEvent: &gateway.MessageCreateEvent{
						Message: discord.Message{Content: c.in},
					},
				})

				actual, err := p.Parse()
				require.NoError(t, err)
				assert.Equal(t, &c.expect, actual)
			})
		}
	})

	failureCases := []struct {
		name   string
		in     string
		expect func(*event.MessageCreate) error
	}{
		{
			name: "only title",
			in:   "abc",
			expect: func(e *event.MessageCreate) error {
				return errhandler.NewInfo(e,
					"An idea must have both a title and description or a title and at least one section.")
			},
		},
		{
			name: "title too long",
			in:   strings.Repeat("a", 257),
			expect: func(e *event.MessageCreate) error {
				return errhandler.NewInfo(e, "The title may be no longer than 256 characters.")
			},
		},
		{
			name: "group title too long",
			in: "title\n" +
				"# " + strings.Repeat("a", 257) + "\n* section contents",
			expect: func(e *event.MessageCreate) error {
				return errhandler.NewInfo(e, "Group titles may be no longer than 256 characters.")
			},
		},
		{
			name: "group title empty",
			in: "title\n" +
				"#" +
				"\n* abc",
			expect: func(e *event.MessageCreate) error {
				return errhandler.NewInfo(e, "A group title may not be empty.")
			},
		},
		{
			name: "empty group",
			in: "title\n" +
				"# abc\n",
			expect: func(e *event.MessageCreate) error {
				return errhandler.NewInfo(e, "A group needs at least one section.")
			},
		},
		{
			name: "adjacent empty group",
			in: "title\n" +
				"# abc\n" +
				"# def",
			expect: func(e *event.MessageCreate) error {
				return errhandler.NewInfo(e, "A group needs at least one section.")
			},
		},
		{
			name: "group title not followed by section",
			in: "title\n" +
				"# abc\n" +
				"def",
			expect: func(e *event.MessageCreate) error {
				return errhandler.NewInfo(e, "I expected the line after the 1st group to start with a section (`*`),"+
					" but I found `d`.")
			},
		},
		{
			name: "section content too long",
			in: "title\n" +
				"* " + strings.Repeat("a", 1025),
			expect: func(e *event.MessageCreate) error {
				return errhandler.NewInfo(e, "Sections may be no longer than 1024 characters.")
			},
		},
	}

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		for _, c := range failureCases {
			c := c
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				e := &event.MessageCreate{
					MessageCreateEvent: &gateway.MessageCreateEvent{
						Message: discord.Message{Content: c.in},
					},
				}

				p := newParser(e)

				_, actual := p.Parse()
				assert.Equal(t, c.expect(e), actual)
			})
		}
	})
}
