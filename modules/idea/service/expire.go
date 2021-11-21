package service

import (
	"context"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/utils/discorderr"
	"github.com/mavolin/adam/pkg/utils/duration"
	"github.com/mavolin/disstate/v4/pkg/state"
	"go.uber.org/multierr"

	"github.com/mavolin/stormy/modules/idea/repository"
	"github.com/mavolin/stormy/modules/idea/service/format"
)

func (service *Service) countExpiredVotes(s *state.State, t time.Time) error {
	service.log.Info("checking if any vote counts are past due")

	cursor, err := service.repo.ExpiredIdeas(t)
	if err != nil {
		return nil
	}
	if cursor == nil { // no expired ideas
		service.log.Info("found no past due votes")
		return nil
	}

	service.log.Infof("found %d+ past due votes", cursor.BatchLength())

	var (
		merr error
		mut  sync.Mutex

		wg sync.WaitGroup
	)

	work := make(chan *repository.Idea)
	numWorkers := cursor.BatchLength()
	if numWorkers > 10_000 {
		numWorkers = 10_000
	}

	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			for idea := range work {
				if err := service.countExpiredVote(s, idea); err != nil {
					mut.Lock()
					merr = multierr.Append(merr, err)
					mut.Unlock()
				}
			}

			wg.Done()
		}()
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		idea, err := cursor.Next(ctx)
		cancel()
		if err != nil {
			return err
		} else if idea == nil {
			break
		}

		work <- idea
	}

	close(work)

	err = service.repo.DeleteExpiredIdeas(t)

	wg.Wait()

	return multierr.Append(merr, err)
}

func (service *Service) countExpiredVote(s *state.State, i *repository.Idea) error {
	service.log.With(
		"guild_id", i.GuildID,
		"channel_id", i.ChannelID,
		"message_id", i.MessageID,
		"vote_until", i.VoteUntil,
	).Debug("counting past due vote")

	// reactions aren't updated by the state, use client
	msg, err := s.Client.Message(i.ChannelID, i.MessageID)
	if err != nil {
		if discorderr.Is(discorderr.As(err), discorderr.UnknownResource...) {
			return nil
		}

		return errors.WithStack(err)
	}

	if len(msg.Embeds) == 0 {
		return nil
	}

	e := msg.Embeds[0]

	delay := duration.Format(time.Since(e.Timestamp.Time()))

	d := format.Votes(i, msg)
	d.RatingField.Value += "\n\nI was offline during the voting deadline. " +
		"The results you see here are from " + delay + " after the deadline."

	// delete the deadline notice, it's not relevant anymore
	e.Footer = new(discord.EmbedFooter)
	e.Timestamp = discord.Timestamp{}
	e.Color = d.Color

	e.Fields = append(e.Fields, d.RatingField)

	if d.SectionVotesField != nil {
		e.Fields = append(e.Fields, *d.SectionVotesField)
	}

	_, err = s.EditEmbeds(i.ChannelID, i.MessageID, e)
	return errors.WithStack(err)
}
