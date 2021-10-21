package wizard

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/msgbuilder"
	"github.com/mavolin/disstate/v4/pkg/state"
)

type Wizard struct {
	s   *state.State
	ctx *plugin.Context

	i     int
	steps []Step
}

type Step struct {
	Question  func(s *state.State, ctx *plugin.Context) *msgbuilder.Builder
	WaitFor   time.Duration
	Validator func(s *state.State, ctx *plugin.Context) error
}

func New(s *state.State, ctx *plugin.Context) *Wizard {
	w := &Wizard{s: s, ctx: ctx}
	ctx.Set(wizardKey{}, w)

	return w
}

func (w *Wizard) AddStep(s Step) {
	w.steps = append(w.steps, s)
}

const maxTries = 3

func (w *Wizard) Start() error {
	for w.i = 0; w.i < len(w.steps); w.i++ {
		step := w.steps[w.i]

		var cancel bool

		q := step.Question(w.s, w.ctx).
			WithAwaitedComponent(msgbuilder.NewActionRow(&cancel).
				With(msgbuilder.NewButton(discord.DangerButton, "Cancel", true)))

		if _, err := q.Reply(); err != nil {
			return err
		}

		for tries := 1; tries <= maxTries; tries++ {
			err := q.Await(step.WaitFor, false)
			if err != nil {
				if err := q.DisableComponents(); err != nil {
					w.ctx.HandleErrorSilently(err)
				}

				return err
			}

			if cancel {
				if err := q.DisableComponents(); err != nil {
					w.ctx.HandleErrorSilently(err)
				}

				return errors.Abort
			}

			if step.Validator == nil {
				break
			}

			err = step.Validator(w.s, w.ctx)
			if err == nil {
				break
			}

			var rerr *RetryError
			if tries < maxTries && errors.As(err, &rerr) {
				w.ctx.HandleError(rerr)
			} else {
				return err
			}
		}

		if err := q.DisableComponents(); err != nil {
			w.ctx.HandleErrorSilently(err)
		}
	}

	return nil
}

type wizardKey struct{}

func Title(ctx *plugin.Context, name string) string {
	w := ctx.Get(wizardKey{}).(*Wizard)

	return fmt.Sprintf("%s (%d/%d)", name, w.i+1, len(w.steps))
}
