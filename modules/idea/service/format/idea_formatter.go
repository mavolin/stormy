// Package format provides means to parse and format ideas.
package format

import (
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/disstate/v4/pkg/event"

	"github.com/mavolin/stormy/internal/errhandler"
	"github.com/mavolin/stormy/modules/idea/repository"
)

// Emojis are the emojis that are used for sections.
var Emojis = []string{"🇦", "🇧", "🇨", "🇩", "🇪", "🇫", "🇬", "🇭", "🇮", "🇯", "🇰", "🇱", "🇲", "🇳", "🇴"}

type ideaFormatter struct {
	set *repository.ChannelSettings
	e   *event.MessageCreate
	i   *parserIdea
	b   strings.Builder

	voteUntil   *time.Time
	numSections int
}

type IdeaData struct {
	Embed    *discord.Embed
	RepoIdea *repository.Idea
}

func newIdeaFormatter(i *parserIdea, e *event.MessageCreate, set *repository.ChannelSettings) *ideaFormatter {
	return &ideaFormatter{i: i, e: e, set: set}
}

func (f *ideaFormatter) format() (*IdeaData, error) {
	embed, err := f.createEmbed()
	if err != nil {
		return nil, err
	}

	return &IdeaData{
		Embed:    embed,
		RepoIdea: f.createRepoIdea(),
	}, nil
}

func (f *ideaFormatter) createRepoIdea() *repository.Idea {
	offset := len(f.i.sections)

	groups := make([]repository.SectionGroup, len(f.i.groups))
	for i, g := range f.i.groups {
		groups[i] = repository.SectionGroup{
			Title:  g.title,
			Emojis: Emojis[offset : offset+len(g.sections)],
		}

		offset += len(g.sections)
	}

	return &repository.Idea{
		GlobalSectionEmojis: Emojis[:len(f.i.sections)],
		Groups:              groups,
		VoteType:            f.set.VoteType,
		VoteUntil:           f.voteUntil,
	}
}

func (f *ideaFormatter) createEmbed() (e *discord.Embed, err error) {
	e = &discord.Embed{Title: f.i.title, Color: f.set.Color}

	if !f.set.Anonymous {
		e.Author = &discord.EmbedAuthor{
			Name: f.e.Author.Username,
			Icon: f.e.Author.AvatarURL(),
		}
	}

	if f.set.VoteDuration > 0 {
		e.Footer = &discord.EmbedFooter{Text: "Voting ends:"}

		voteUntil := f.e.Timestamp.Time().Add(f.set.VoteDuration)

		f.voteUntil = &voteUntil
		e.Timestamp = discord.NewTimestamp(voteUntil)
	}

	f.b.Grow(4096) // highest len max is 4096, the max len of a description

	e.Description, err = f.description()
	if err != nil {
		return nil, err
	}

	e.Fields, err = f.groups()
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (f *ideaFormatter) description() (string, error) {
	if f.i.desc != "" {
		f.b.WriteString(f.i.desc)
		f.b.WriteString("\n\n")
	}

	f.sections(f.i.sections)

	if f.b.Len() > 4096 {
		return "", errhandler.NewInfof(f.e,
			"The description and titleless sections are a total of %d characters too long. Reduce their length "+
				"or assign titles to some of the sections.", f.b.Len()-4096)
	}

	return f.b.String(), nil
}

func (f *ideaFormatter) groups() ([]discord.EmbedField, error) {
	if len(f.i.groups) == 0 {
		return nil, nil
	}

	fields := make([]discord.EmbedField, len(f.i.groups))

	for i, ts := range f.i.groups {
		f.b.Reset()

		f.sections(ts.sections)
		if f.b.Len() > 1024 {
			return nil, errhandler.NewInfof(f.e,
				"The sections in the group with the name %s are a total of %d characters too long.",
				ts.title, f.b.Len()-1024)
		}

		fields[i] = discord.EmbedField{Name: ts.title, Value: f.b.String()}
	}

	return fields, nil
}

// sections formats a slice of sections.
// Unlike other methods, sections does not reset the builder.
func (f *ideaFormatter) sections(ss []string) {
	for i, s := range ss {
		if i > 0 {
			f.b.WriteRune('\n')
		}

		emoji := Emojis[f.numSections]

		f.b.WriteString(emoji)
		f.b.WriteRune(' ')
		f.b.WriteString(s)

		f.numSections++
	}
}
