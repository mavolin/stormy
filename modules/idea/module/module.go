// Package module provides the idea module.
package module

import (
	"github.com/mavolin/adam/pkg/impl/module"
	"github.com/mavolin/adam/pkg/plugin"

	"github.com/mavolin/stormy/modules/idea/module/disable"
	"github.com/mavolin/stormy/modules/idea/module/how"
	"github.com/mavolin/stormy/modules/idea/module/setup"
	"github.com/mavolin/stormy/modules/idea/repository"
)

// New creates a new module that provides all idea related commands.
func New(r repository.Repository) plugin.Module {
	mod := module.New(module.Meta{
		Name:             "idea",
		ShortDescription: "Brainstorm ideas in a channel.",
	})

	mod.AddCommand(disable.New(r))
	mod.AddCommand(how.New())
	mod.AddCommand(setup.New(r))

	return mod
}
