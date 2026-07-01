package log

import (
	"context"

	"github.com/rs/zerolog"
)

func Debug(ctx context.Context, msg string) {
	Ctx(ctx).Debug().Caller(1).Msg(msg)
}

func Info(ctx context.Context, msg string) {
	Ctx(ctx).Info().Caller(1).Msg(msg)
}

func Warn(ctx context.Context, msg string) {
	Ctx(ctx).Warn().Caller(1).Msg(msg)
}

func Error(ctx context.Context, msg string) {
	Ctx(ctx).Error().Caller(1).Msg(msg)
}

func Infow(ctx context.Context, msg string, fields ...Field) {
	InfowDepth(ctx, 2, msg, fields...)
}

func InfowDepth(ctx context.Context, depth int, msg string, fields ...Field) {
	applyFields(Ctx(ctx).Info(), fields...).Caller(depth).Msg(msg)
}

func Warnw(ctx context.Context, msg string, fields ...Field) {
	WarnwDepth(ctx, 2, msg, fields...)
}

func WarnwDepth(ctx context.Context, depth int, msg string, fields ...Field) {
	applyFields(Ctx(ctx).Warn(), fields...).Caller(depth).Msg(msg)
}

func Errorw(ctx context.Context, msg string, fields ...Field) {
	ErrorwDepth(ctx, 2, msg, fields...)
}

func ErrorwDepth(ctx context.Context, depth int, msg string, fields ...Field) {
	applyFields(Ctx(ctx).Error(), fields...).Caller(depth).Msg(msg)
}

func Errorwf(ctx context.Context, err error, fields ...Field) {
	ErrorwfDepth(ctx, 2, err, fields...)
}

func ErrorwfDepth(ctx context.Context, depth int, err error, fields ...Field) {
	applyFields(Ctx(ctx).Error().Err(err), fields...).Caller(depth).Msg("error")
}

func applyFields(e *zerolog.Event, fields ...Field) *zerolog.Event {
	for _, f := range fields {
		f(e)
	}
	return e
}
