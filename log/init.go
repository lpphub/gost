package log

import (
	"context"

	"github.com/rs/zerolog"
)

var std *zerolog.Logger

func Init(opts ...Option) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	l := newZerolog(cfg)
	std = &l
}

func L() *zerolog.Logger {
	return std
}

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

func Errorw(ctx context.Context, err error) {
	Ctx(ctx).Error().Caller(1).Err(err).Msg("error")
}

func Infof(ctx context.Context, msg string, fields ...Field) {
	applyFields(Ctx(ctx).Info(), fields...).Caller(1).Msg(msg)
}

func Warnf(ctx context.Context, msg string, fields ...Field) {
	applyFields(Ctx(ctx).Warn(), fields...).Caller(1).Msg(msg)
}

func Errorf(ctx context.Context, msg string, fields ...Field) {
	applyFields(Ctx(ctx).Error(), fields...).Caller(1).Msg(msg)
}

func Errorwf(ctx context.Context, err error, fields ...Field) {
	applyFields(Ctx(ctx).Error().Err(err), fields...).Caller(1).Msg("error")
}

func applyFields(e *zerolog.Event, fields ...Field) *zerolog.Event {
	for _, f := range fields {
		f(e)
	}
	return e
}
