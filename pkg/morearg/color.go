package morearg

import (
	"regexp"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/i18n"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/state"
)

// Color is the type used to retrieve discord.Colors.
var Color plugin.ArgType = new(color)

type color struct{}

func (c *color) GetName(*i18n.Localizer) string {
	return "Color"
}

func (c *color) GetDescription(*i18n.Localizer) string {
	return "A hexadecimal color such as `44c5f3`."
}

var colorRegexp = regexp.MustCompile("^[a-fA-F0-9]{6}$")

func (c *color) Parse(_ *state.State, ctx *plugin.ParseContext) (interface{}, error) {
	if !colorRegexp.MatchString(ctx.Raw) {
		return nil, plugin.NewArgumentError("`" + ctx.Raw + "` is not a valid color.")
	}

	cInt, err := strconv.ParseInt(ctx.Raw, 16, 32)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse valid color")
	}

	return discord.Color(cInt), nil
}

func (c *color) GetDefault() interface{} {
	return discord.Color(0)
}
