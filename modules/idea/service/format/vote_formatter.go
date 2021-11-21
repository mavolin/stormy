package format

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/dustin/go-humanize"

	"github.com/mavolin/stormy/internal/stdcolor"
	"github.com/mavolin/stormy/modules/idea/repository"
)

type voteFormatter struct {
	i   *repository.Idea
	msg *discord.Message

	percentRating float64
	counts        map[string]int

	b strings.Builder
}

func newVoteFormatter(i *repository.Idea, msg *discord.Message) *voteFormatter {
	f := &voteFormatter{i: i, msg: msg}
	f.b.Grow(1024) // length of a field

	return f
}

type VoteData struct {
	RatingField discord.EmbedField
	Color       discord.Color
}

func (f *voteFormatter) format() (d VoteData) {
	f.calcCounts()

	f.rating()
	if len(f.i.GlobalSectionEmojis) != 0 || len(f.i.Groups) != 0 {
		f.sectionVotes()
	}

	d.RatingField = discord.EmbedField{
		Name:  "━━━━━━━━━━",
		Value: f.b.String(),
	}

	switch {
	case f.percentRating < 0.333:
		d.Color = stdcolor.Red
	case f.percentRating < 0.666:
		d.Color = stdcolor.Yellow
	default:
		d.Color = stdcolor.Green
	}

	return d
}

func (f *voteFormatter) calcCounts() {
	es := f.i.VoteType.Emojis()

	f.counts = make(map[string]int, len(Emojis)+len(es))

Votes:
	for _, e := range es {
		for _, r := range f.msg.Reactions {
			if r.Emoji.IsUnicode() && r.Emoji.Name == e {
				f.counts[e] = r.Count - 1 // subtract our own reaction
				continue Votes
			}
		}
	}

GlobalSections:
	for _, e := range f.i.GlobalSectionEmojis {
		for _, r := range f.msg.Reactions {
			if r.Emoji.IsUnicode() && r.Emoji.Name == e {
				f.counts[e] = r.Count - 1 // subtract our own reaction
				continue GlobalSections
			}
		}
	}

	for _, g := range f.i.Groups {
	GroupSections:
		for _, e := range g.Emojis {
			for _, r := range f.msg.Reactions {
				if r.Emoji.IsUnicode() && r.Emoji.Name == e {
					f.counts[e] = r.Count - 1 // subtract our own reaction
					continue GroupSections
				}
			}
		}
	}
}

func (f *voteFormatter) rating() {
	f.b.WriteString("**Overall Rating:** ")

	es := f.i.VoteType.Emojis()
	for _, e := range es {
		_, ok := f.counts[e]
		if !ok {
			f.b.WriteString("*Could not calculate as some reactions were deleted*")
		}
	}

	if len(es) == 2 {
		f.twoEmojiRating(es[0], es[1])
		return
	}

	f.weightedRating(es)
}

func (f *voteFormatter) twoEmojiRating(posEmoji, negEmoji string) {
	posCount := f.counts[posEmoji]
	negCount := f.counts[negEmoji]

	if posCount+negCount <= 0 {
		f.b.WriteString("-/- %")
		return
	}

	// get the average in percent
	rating := float64(posCount) / float64(posCount+negCount)

	f.percentRating = rating

	rating *= 1000 // scale from 0 to 1000
	rating = math.Round(rating)
	rating /= 10 // scale from 0 to 100

	f.b.WriteString(fmt.Sprintf("%.1f%%\n", rating))

	f.b.WriteString(posEmoji)
	f.b.WriteString(": ")
	f.b.WriteString(strconv.Itoa(posCount))
	f.b.WriteString("x  ")

	f.b.WriteString(negEmoji)
	f.b.WriteString(": ")
	f.b.WriteString(strconv.Itoa(negCount))
	f.b.WriteString("x")
}

func (f *voteFormatter) weightedRating(emojis []string) {
	var totalVotes int
	var weightedVotes int

	for i, e := range emojis {
		counts := f.counts[e]

		totalVotes += counts
		weightedVotes += (len(emojis) - i) * counts
	}

	// get the average rating from 1 to len(emojis)
	rating := float64(weightedVotes) / float64(totalVotes)

	f.percentRating = rating * (100 / float64(len(emojis)))

	// shift the one decimal place to the right, i.e. 5 becomes 50
	rating *= 10
	rating = math.Round(rating)

	// shift back
	rating /= 10

	f.b.WriteString(fmt.Sprintf("%.1f/%d\n", rating, len(emojis)))

	for i, e := range emojis {
		if i > 0 {
			f.b.WriteString("  ")
		}

		f.b.WriteString(e)
		f.b.WriteString(": ")
		f.b.WriteString(strconv.Itoa(f.counts[e]))
		f.b.WriteRune('x')
	}
}

func (f *voteFormatter) sectionVotes() {
	f.b.WriteString("\n\n")

	for i, e := range f.i.GlobalSectionEmojis {
		if i > 0 {
			f.b.WriteString("  ")
		}

		f.b.WriteString(e)
		f.b.WriteString(": ")

		count, ok := f.counts[e]
		if !ok {
			f.b.WriteString("*Reaction was removed*")
			continue
		}

		f.b.WriteString(humanize.Comma(int64(count)))
		f.b.WriteRune('x')
	}

	for i, g := range f.i.Groups {
		if i > 0 || len(f.i.GlobalSectionEmojis) > 0 {
			f.b.WriteString("\n")
		}

		f.b.WriteString("**")
		f.b.WriteString(g.Title)
		f.b.WriteString("**\n")

		for j, e := range g.Emojis {
			if j > 0 {
				f.b.WriteString("  ")
			}

			f.b.WriteString(e)
			f.b.WriteString(": ")

			count, ok := f.counts[e]
			if !ok {
				f.b.WriteString("*Reaction was removed*")
				continue
			}

			f.b.WriteString(humanize.Comma(int64(count)))
			f.b.WriteString("x")
		}
	}
}
