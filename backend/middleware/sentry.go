package middleware

import (
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/rs/zerolog/log"
)

func CreateNewSentryMiddleware() *sentryhttp.Handler {
	// Init Sentry first
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:                "https://1e36660a7a5e022a8f602034da95eca7@o4506142641356800.ingest.sentry.io/4506142665277440",
		EnableTracing:      true,
		TracesSampleRate:   1.0,
		ProfilesSampleRate: 1.0,
	}); err != nil {
		log.Error().Err(err).Msg("Sentry initialization failed")
	}
	defer sentry.Flush(time.Second)
	log.Debug().Msg("sentry initialized")

	return sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
}
