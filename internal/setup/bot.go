package setup

import (
	"bytes"
	"io"
	"reflect"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/httputil/httpdriver"
	"github.com/diamondburned/arikawa/v3/utils/wsutil"
	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/i18n"
	"github.com/mavolin/adam/pkg/impl/command/help"
	"github.com/mavolin/disstate/v4/pkg/state"
	"github.com/mavolin/sentryadam/pkg/sentryadam"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mavolin/stormy/internal/errhandler"
	"github.com/mavolin/stormy/internal/stdcolor"
	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/pkg/utils/zapadam"
	"github.com/mavolin/stormy/pkg/utils/zapstate"
	"github.com/mavolin/stormy/service/idea"
)

type BotOptions struct {
	Token        string
	Owners       []discord.UserID
	Status       discord.Status
	ActivityType discord.ActivityType
	ActivityName string
	ActivityURL  discord.URL

	Logger *zap.SugaredLogger
	Hub    *sentry.Hub

	Repository repository.Repository
}

func Bot(o BotOptions) (*bot.Bot, error) {
	o.Logger = o.Logger.Named("bot")

	b, err := bot.New(bot.Options{
		Token:                o.Token,
		Owners:               o.Owners,
		Status:               o.Status,
		ActivityType:         o.ActivityType,
		ActivityName:         o.ActivityName,
		ActivityURL:          o.ActivityURL,
		GatewayErrorHandler:  errhandler.NewGatewayHandler(o.Logger, o.Hub),
		StatePanicHandler:    errhandler.NewStatePanicHandler(o.Logger, o.Hub),
		NoDefaultMiddlewares: true,
	})
	if err != nil {
		return nil, err
	}

	b.State.ErrorHandler = errhandler.NewStateErrorHandler(b.State, o.Logger, o.Hub)

	configureErrors()

	if o.Logger.Desugar().Core().Enabled(zapcore.DebugLevel) {
		addDebugLoggers(b, o.Logger)
	}

	addMiddlewares(b, o)
	addPlugins(b, o.Repository)

	b.AddIntents(b.State.DeriveIntents())
	b.AddIntents(gateway.IntentGuildMessageTyping)

	return b, err
}

func addMiddlewares(b *bot.Bot, o BotOptions) {
	// ================================ State Middlewares ================================
	b.State.AddMiddleware(zapstate.NewMiddleware(o.Logger))

	// ================================ Bot Middlewares ================================

	sentryWrapper := sentryadam.New(sentryadam.Options{Hub: o.Hub})

	b.AddMiddleware(sentryWrapper.PreRouteMiddleware)
	b.AddMiddleware(zapadam.NewFallbackMiddleware(o.Logger))
	b.AddMiddleware(bot.CheckMessageType)
	b.AddMiddleware(bot.CheckHuman)
	b.AddMiddleware(bot.NewSettingsRetriever(bot.StaticSettings()))
	b.AddMiddleware(bot.CheckPrefix)
	b.AddMiddleware(bot.FindCommand)
	b.AddMiddleware(bot.CheckChannelTypes)
	b.AddMiddleware(bot.CheckBotPermissions)
	b.AddMiddleware(bot.NewThrottlerChecker(bot.DefaultThrottlerErrorCheck))
	b.AddMiddleware(sentryWrapper.PostRouteMiddleware)

	b.AddMiddleware(zapadam.NewMiddleware(o.Logger))

	b.AddPostMiddleware(bot.CheckRestrictions)
	b.AddPostMiddleware(bot.ParseArgs)
	b.AddPostMiddleware(sentryWrapper.PreInvokeMiddleware)
	b.AddPostMiddleware(bot.InvokeCommand)
}

func addPlugins(b *bot.Bot, r repository.Repository) {
	b.AddCommand(help.New(help.Options{}))

	b.AddModule(idea.NewModule(r))
}

func configureErrors() {
	errors.Log = errhandler.ErrorLog

	errors.SetErrorEmbedTemplate(func(*i18n.Localizer) discord.Embed {
		return discord.Embed{Title: "Err", Color: stdcolor.Red}
	})

	errors.SetInfoEmbedTemplate(func(*i18n.Localizer) discord.Embed {
		return discord.Embed{Title: "Info", Color: stdcolor.Default}
	})
}

func addDebugLoggers(b *bot.Bot, l *zap.SugaredLogger) {
	wsutil.WSDebug = l.Named("ws_debug").Debug

	reqLogger := l.Named("api_client")
	b.State.Client.Client.OnRequest = append(b.State.Client.Client.OnRequest, func(r httpdriver.Request) (err error) {
		dr, ok := r.(*httpdriver.DefaultRequest)
		if !ok {
			return nil
		}

		var body []byte

		if dr.Body != nil {
			body, err = io.ReadAll(dr.Body)
			if err != nil {
				return err
			}

			dr.Body = io.NopCloser(bytes.NewReader(body))
		}

		reqLogger.With(
			"url", dr.URL.String(),
			"method", dr.Method,
			"header", dr.Header,
			"body", body,
		).Debug("making a request to ", dr.URL.Host)

		return nil
	})

	ehLogger := l.Named("event_handler")

	b.State.AddHandler(func(_ *state.State, e interface{}) {
		t := reflect.TypeOf(e)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		ehLogger.With(
			"type", t.Name(),
			"data", e,
		).Debug("received gateway event")
	})
}
