package idea

import (
	"github.com/mavolin/adam/pkg/impl/module"
	"github.com/mavolin/adam/pkg/plugin"

	"github.com/mavolin/stormy/plugin/idea/disable"
	"github.com/mavolin/stormy/plugin/idea/how"
	"github.com/mavolin/stormy/plugin/idea/setup"
)

type Repository interface {
	disable.Repository
	setup.Repository
}

func New(r Repository) plugin.Module {
	mod := module.New(module.Meta{
		Name:             "idea",
		ShortDescription: "Brainstorm ideas in a channel.",
	})

	mod.AddCommand(disable.New(r))
	mod.AddCommand(how.New())
	mod.AddCommand(setup.New(r))

	return mod
}
