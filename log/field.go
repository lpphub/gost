package log

import (
	"time"

	"github.com/rs/zerolog"
)

type Field func(*zerolog.Event)

type CtxField func(zerolog.Context) zerolog.Context

func Str(key, val string) Field {
	return func(e *zerolog.Event) { e.Str(key, val) }
}

func Int64(key string, val int64) Field {
	return func(e *zerolog.Event) { e.Int64(key, val) }
}

func Float64(key string, val float64) Field {
	return func(e *zerolog.Event) { e.Float64(key, val) }
}

func Dur(key string, val time.Duration) Field {
	return func(e *zerolog.Event) { e.Dur(key, val) }
}

func Err(err error) Field {
	return func(e *zerolog.Event) { e.Err(err) }
}

func Any(key string, val any) Field {
	return func(e *zerolog.Event) { e.Any(key, val) }
}


func CtxStr(key, val string) CtxField {
	return func(c zerolog.Context) zerolog.Context { return c.Str(key, val) }
}

func CtxInt64(key string, val int64) CtxField {
	return func(c zerolog.Context) zerolog.Context { return c.Int64(key, val) }
}

func CtxDur(key string, val time.Duration) CtxField {
	return func(c zerolog.Context) zerolog.Context { return c.Dur(key, val) }
}
