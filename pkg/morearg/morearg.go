package morearg

import "github.com/mavolin/adam/pkg/plugin"

type undefined struct{}

// Undefined is a placeholder for the default value, that can be used if a
// distinction omitted and present, but default value is necessary.
var Undefined = undefined{}

func FlagOrDefault(ctx *plugin.Context, name string, def interface{}) interface{} {
	val := ctx.Flags[name]
	if val == Undefined {
		return def
	}

	return val
}
