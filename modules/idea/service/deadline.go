package service

import (
	"container/list"
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/modules/idea/repository"
	"github.com/mavolin/stormy/modules/idea/service/format"
)

// =============================================================================
// Service
// =====================================================================================

func (service *Service) countVotes(s *state.State, i *repository.Idea) {
	err := service.sf.DoSync(i.MessageID, func() error {
		service.log.With(
			"message_id", i.MessageID,
			"channel_id", i.ChannelID,
			"guild_id", i.GuildID,
		).Info("counting votes")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := service.deleteIdea(ctx, i.MessageID); err != nil {
			return errors.WithStack(err)
		}

		// reactions aren't cached by the state, get the message using the
		// client
		msg, err := s.Client.Message(i.ChannelID, i.MessageID)
		if err != nil {
			return errors.WithStack(err)
		}

		if len(msg.Embeds) == 0 {
			return nil
		}

		e := msg.Embeds[0]

		d := format.Votes(i, msg)

		e.Footer = &discord.EmbedFooter{Text: "Voting ended:"}
		e.Color = d.Color

		e.Fields = append(e.Fields, d.RatingField)

		_, err = s.EditEmbeds(i.ChannelID, i.MessageID, e)
		return errors.WithStack(err)
	})

	service.errhandler.Capture(err)
}

// =============================================================================
// deadlineWatcher
// =====================================================================================

const maxDeadlineQueueLen = 50

type deadlineWatcher struct {
	service *Service
	state   *state.State

	r repository.IdeaRepository

	// queue is a queue of expiring ideas.
	queue *list.List

	ic chan *repository.Idea
}

func newDeadlineWatcher(service *Service, state *state.State, r repository.IdeaRepository) *deadlineWatcher {
	return &deadlineWatcher{
		service: service,
		state:   state,
		r:       r,
		queue:   list.New(),
		ic:      make(chan *repository.Idea),
	}
}

func (w *deadlineWatcher) start(t time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ideas, err := w.r.ExpiringIdeas(ctx, t, 0, maxDeadlineQueueLen)
	if err != nil {
		return err
	}

	for _, i := range ideas {
		w.queue.PushBack(i)
	}

	go w.loop()

	return nil
}

func (w *deadlineWatcher) addIdea(i *repository.Idea) {
	w.ic <- i
}

func (w *deadlineWatcher) loop() {
	var t *time.Timer
	if w.queue.Len() > 0 {
		frontVal := w.queue.Front().Value.(*repository.Idea)
		t = time.NewTimer(time.Until(*frontVal.VoteUntil))
	} else {
		// create an empty timer; hacky, but whatever
		t = time.NewTimer(0)
		<-t.C
	}

	for {
		select {
		case idea := <-w.ic:
			// The only case in which t.C needs to be drained is if the timer
			// is still running, which is only the case if we have elements
			// in the queue.
			// If there are no elements in the queue, t is already stopped.
			if !t.Stop() && w.queue.Len() > 0 {
				<-t.C
			}

			w.insertIdea(idea)

			frontVal := w.queue.Front().Value.(*repository.Idea)

			t.Reset(time.Until(*frontVal.VoteUntil))
		case <-t.C:
			front := w.queue.Front()

			frontVal := front.Value.(*repository.Idea)
			go w.service.countVotes(w.state, frontVal)

			w.queue.Remove(front)

			if w.queue.Len() != 0 {
				frontVal = w.queue.Front().Value.(*repository.Idea)
				t.Reset(time.Until(*frontVal.VoteUntil))
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

				ideas, err := w.r.ExpiringIdeas(ctx, *frontVal.VoteUntil, frontVal.MessageID, maxDeadlineQueueLen)
				if err != nil {
					w.service.errhandler.Capture(err)
				}

				cancel()

				for _, i := range ideas {
					w.queue.PushBack(i)
				}
			}
		}
	}
}

// =============================================================================
// Utils
// =====================================================================================

func (w *deadlineWatcher) insertIdea(i *repository.Idea) {
	if w.queue.Len() == 0 {
		w.queue.PushBack(i)
		return
	}

	elem := w.queue.Front()

Elems:
	for elem != nil {
		ev := elem.Value.(*repository.Idea)

		switch {
		case ev.VoteUntil.Before(*i.VoteUntil):
			fallthrough
		case ev.VoteUntil.Equal(*i.VoteUntil) && ev.MessageID < i.MessageID:
			elem = elem.Next()
		default:
			break Elems
		}
	}

	if elem == nil {
		return
	}

	w.queue.InsertBefore(i, elem)

	if w.queue.Len() > maxDeadlineQueueLen {
		w.queue.Remove(w.queue.Back())
	}
}
