package format

import (
	"github.com/dustin/go-humanize"
	"github.com/mavolin/disstate/v4/pkg/event"

	"github.com/mavolin/stormy/internal/errhandler"
)

type (
	parser struct {
		e *event.MessageCreate

		i parserIdea

		raw        []rune
		start, pos int

		sectionNum int

		state state
	}

	state func() (state, error)
)

type (
	parserIdea struct {
		// title is the title of the idea.
		title string
		// desc is the optional description of the idea.
		desc string
		// sections contains all nameless sections in the order they appear in.
		sections []string
		// groups contains all groups in the order in which they first
		// appeared.
		groups []group
	}

	group struct {
		title    string
		sections []string
	}
)

// newParser creates a new *parser.
func newParser(e *event.MessageCreate) *parser {
	l := &parser{e: e, raw: []rune(e.Content)}
	l.state = l.title

	return l
}

func (p *parser) Parse() (_ *parserIdea, err error) {
	for p.state != nil {
		p.state, err = p.state()
		if err != nil {
			return nil, err
		}
	}

	return &p.i, nil
}

// =============================================================================
// Helpers
// =====================================================================================

// has checks if there are at least min runes remaining.
func (p *parser) has(min int) bool {
	return p.pos+min <= len(p.raw)
}

func (p *parser) next() rune {
	if !p.has(1) {
		return 0
	}

	p.pos++
	return p.raw[p.pos-1]
}

func (p *parser) skip() {
	p.pos++
}

// backup goes one character back.
func (p *parser) backup() {
	p.pos--
}

// peek peeks n characters ahead or behind, without changing the position.
func (p *parser) peek(n int) rune {
	if !p.has(n) {
		return 0
	}

	return p.raw[p.pos+n-1]
}

// ignore ignores all content up to this point.
// It starts at the upcoming character.
func (p *parser) ignore() {
	p.start = p.pos
}

func (p *parser) ignoreAll(r rune) {
	for p.has(1) {
		if p.next() != r {
			p.backup()
			break
		}
	}

	p.ignore()
}

func (p *parser) consumeUntil(seqs ...[]rune) {
	var nlPos int

Loop:
	for p.has(1) {
		c := p.next()
		if c == '\n' && nlPos <= 0 {
			nlPos = p.pos - 1
		}

	SeqLoop:
		for _, seq := range seqs {
			if c == seq[0] {
				for i, r := range seq[1:] {
					if p.peek(i+1) != r {
						continue SeqLoop
					}
				}

				p.backup()
				break Loop
			}
		}

		if c != '\n' {
			nlPos = 0
		}
	}

	if nlPos > 0 {
		p.pos = nlPos
	}
}

func (p *parser) consumeRange(min, max rune) {
	for p.has(1) {
		c := p.next()
		if c < min || c > max {
			p.backup()
			break
		}
	}
}

func (p *parser) get() string {
	content := string(p.raw[p.start:p.pos])
	p.start = p.pos

	return content
}

func (p *parser) len() int {
	return p.pos - p.start
}

// =============================================================================
// States
// =====================================================================================

func (p *parser) title() (state, error) {
	p.consumeUntil([]rune{'\n'})

	p.i.title = p.get()

	// This purposefully compares the number of bytes of the title with the
	// length limit for titles, instead of using parser.len() which returns the
	// number of runes.
	//
	// While Discord does count the runes instead of bytes in some cases, they
	// aren't consistent with this behavior.
	// For example, they count (at least some) emojis as half of their bytes,
	// e.g. 😀, which has 4 bytes, is counted as 2 runes, and 🇩🇪, which has 8
	// bytes, is counted as 4 runes, even though it is only 2.
	// I have no clue if there are other exceptions, but until this is
	// consistent with the actual rune count, I'm using the byte count to
	// be safe, i.e. getting length limit errors.
	// Everyone who needs longer titles, go complain to Discord.
	//
	// The same of course applies to all other bounds checks in here.
	if len(p.i.title) > 256 {
		return nil, errhandler.NewInfo(p.e, "The title may be no longer than 256 characters.")
	}

	return p.descriptionOrSection, nil
}

func (p *parser) descriptionOrSection() (state, error) {
	p.ignoreAll('\n')

	if !p.has(1) {
		return nil, errhandler.NewInfo(
			p.e, "An idea must have both a title and description or a title and at least one section.")
	}

	if p.peek(1) == '#' {
		return p.groupTitle, nil
	} else if p.peek(1) == '*' {
		return p.section, nil
	}

	return p.description, nil
}

func (p *parser) description() (state, error) {
	p.consumeUntil([]rune("\n*"), []rune("\n#"))

	p.i.desc = p.get()

	if !p.has(1) {
		return nil, nil
	}

	return p.sectionOrGroupTitle, nil
}

// sectionOrGroupTitle returns either p.groupTitle or p.section.
// When called the next characters must be either a hash or an asterisk,
// optionally preceded by any number of newlines.
func (p *parser) sectionOrGroupTitle() (state, error) {
	p.ignoreAll('\n')

	if p.peek(1) == '#' {
		return p.groupTitle, nil
	}

	return p.section, nil
}

func (p *parser) groupTitle() (state, error) {
	p.skip()
	p.ignore() // ignore the hash

	if !p.has(1) || p.peek(1) == '\n' {
		return nil, errhandler.NewInfo(p.e, "A group title may not be empty.")
	}

	p.ignoreAll(' ')

	p.consumeUntil([]rune{'\n'})

	title := p.get()
	if len(title) > 256 {
		return nil, errhandler.NewInfo(p.e, "Group titles may be no longer than 256 characters.")
	}
	p.i.groups = append(p.i.groups, group{title: title})

	p.ignoreAll('\n')

	if !p.has(1) || p.peek(1) == '#' {
		return nil, errhandler.NewInfo(p.e, "A group needs at least one section.")
	} else if c := p.peek(1); c != '*' {
		return nil, errhandler.NewInfof(p.e, "I expected the line after the %s group to start with a section (`*`),"+
			" but I found `%s`.", humanize.Ordinal(len(p.i.groups)), string(c))
	}

	return p.section, nil
}

func (p *parser) section() (_ state, err error) {
	// skip and ignore the asterisk
	p.skip()
	p.ignore()

	if p.sectionNum >= len(Emojis) {
		return nil, errhandler.NewInfo(p.e, "An idea cannot have more than 15 sections.")
	}

	if !p.has(1) || p.peek(1) == '\n' {
		return nil, errhandler.NewInfo(p.e, "A section may not be empty.")
	}

	p.sectionNum++

	p.ignoreAll(' ')

	p.consumeUntil([]rune("\n*"), []rune("\n#"))

	content := p.get()
	if len(content) > 1024 {
		return nil, errhandler.NewInfo(p.e, "Sections may be no longer than 1024 characters.")
	}

	// content cannot be empty, since discord will trim the whitespace from
	// each line, hence for it to be empty there couldn't have been a

	if len(p.i.groups) > 0 {
		sections := p.i.groups[len(p.i.groups)-1].sections
		sections = append(sections, content)

		p.i.groups[len(p.i.groups)-1].sections = sections
	} else {
		p.i.sections = append(p.i.sections, content)
	}

	if !p.has(1) {
		return nil, nil
	}

	return p.sectionOrGroupTitle, nil
}
