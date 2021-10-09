// Package action provides the action module.
package action

import (
	"github.com/mavolin/adam/pkg/impl/module"
	"github.com/mavolin/adam/pkg/plugin"

	"github.com/mavolin/stormy/pkg/action"
	"github.com/mavolin/stormy/plugin/action/enable"
	"github.com/mavolin/stormy/plugin/action/list"
)

func New(actions ...action.Action) plugin.Module {
	mod := module.New(module.Meta{
		Name:             "action",
		ShortDescription: "Manage actions for a channel.",
		LongDescription:  "Enable, list, or disable actions in the calling channel.",
	})

	mod.AddCommand(enable.New(actions...))
	mod.AddCommand(list.New(actions...))

	return mod
}