// +build !go1.7

package logging

import (
	"context"
	"os"
)

// LogFromContext returns a Logger that is ready to use from the Context provided. In a case where a Logger
// has not been stored in the Context previously (using SaveToContext), LogFromContext will fall back on a
// Logger that writes to stderr, is set to InfoLvl, and has no Sentry configuration. This is to help debug
// Logger configuration errors; in production, SaveToContext should always be used before trying to retrieve
// the Logger wtih LogFromContext. Normally, SaveToContext should be called as part of application startup
// when the Logger is instantiated.
func LogFromContext(c context.Context) Logger {
	ctxVal := c.Value(contextKey)
	if ctxVal == nil {
		logger, err := New(InfoLvl, os.Stderr, "", nil)
		if err != nil {
			panic(err.Error())
		}
		return logger
	}
	logger, ok := ctxVal.(Logger)
	if !ok {
		logger, err := New(InfoLvl, os.Stderr, "", nil)
		if err != nil {
			panic(err.Error())
		}
		return logger
	}
	return logger
}

// SaveToContext adds a Logger to the supplied Context, returning the new Context that contains the Logger.
// SaveToContext should generally be called during application startup, when the Logger is instantiated. Once
// a Logger is stored with SaveToContext, it can be retrieved using LogFromContext.
func SaveToContext(l Logger, base context.Context) context.Context {
	return context.WithValue(base, contextKey, l)
}
