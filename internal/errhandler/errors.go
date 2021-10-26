package errhandler

import (
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/i18n"
	"github.com/mavolin/adam/pkg/utils/discorderr"
	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/mavolin/disstate/v4/pkg/state"
)

type handler interface {
	Handle(s *state.State) error
}

type deleteMessagesKey struct{}

// DeleteError is a middleware that can be added to signal that both the error
// message and the original message shall be deleted 15 seconds after sending
// the error.
func DeleteError(_ *state.State, e *event.MessageCreate) {
	e.Set(deleteMessagesKey{}, struct{}{})
}

func deleteErr(s *state.State, e *event.MessageCreate, errID discord.MessageID) error {
	if e.Get(deleteMessagesKey{}) == nil {
		return nil
	}

	time.Sleep(15 * time.Second)

	err := s.DeleteMessages(e.ChannelID, []discord.MessageID{e.ID, errID}, "")
	if discorderr.Is(discorderr.As(err), discorderr.UnknownResource...) {
		return nil
	}

	return errors.WithStack(err)
}

// =============================================================================
// Info
// =====================================================================================

type Info struct {
	e           *event.MessageCreate
	Description string
}

var (
	_ error   = new(Info)
	_ handler = new(Info)
)

func NewInfo(e *event.MessageCreate, description string) *Info {
	return &Info{e: e, Description: description}
}

func (i *Info) Handle(s *state.State) error {
	e := errors.NewInfoEmbed(i18n.NewFallbackLocalizer())
	e.Description = i.Description

	msg, err := s.SendEmbeds(i.e.ChannelID, e)
	if err != nil {
		return errors.WithStack(err)
	}

	return deleteErr(s, i.e, msg.ID)
}

func (i *Info) Error() string {
	return "errhandler: Info: " + i.Description
}

// =============================================================================
// Error
// =====================================================================================

type Error struct {
	e           *event.MessageCreate
	Description string
}

var (
	_ error   = new(Error)
	_ handler = new(Error)
)

func NewError(e *event.MessageCreate, description string) *Error {
	return &Error{e: e, Description: description}
}

func (e *Error) Handle(s *state.State) error {
	embed := errors.NewErrorEmbed(i18n.NewFallbackLocalizer())
	embed.Description = e.Description

	msg, err := s.SendEmbeds(e.e.ChannelID, embed)
	if err != nil {
		return errors.WithStack(err)
	}

	return deleteErr(s, e.e, msg.ID)
}

func (e *Error) Error() string {
	return "errhandler: Err: " + e.Description
}

// =============================================================================
// InternalError
// =====================================================================================

type InternalError struct {
	e           *event.MessageCreate
	Description string
	Err         error
}

var (
	_ error   = new(InternalError)
	_ handler = new(InternalError)
)

var defaultInternalErrorDesc = errors.NewWithStack("").Description(i18n.NewFallbackLocalizer())

func NewInternalError(e *event.MessageCreate, err error) *InternalError {
	return NewInternalErrorWithDescription(e, err, defaultInternalErrorDesc)
}

func NewInternalErrorWithDescription(e *event.MessageCreate, err error, description string) *InternalError {
	return &InternalError{
		e:           e,
		Description: description,
		Err:         errors.WithStack(err),
	}
}

func (e *InternalError) Handle(s *state.State) error {
	embed := errors.NewErrorEmbed(i18n.NewFallbackLocalizer())
	embed.Description = e.Description

	msg, err := s.SendEmbeds(e.e.ChannelID, embed)
	if err != nil {
		return errors.WithStack(err)
	}

	return deleteErr(s, e.e, msg.ID)
}

func (e *InternalError) Error() string {
	return e.Err.Error()
}
