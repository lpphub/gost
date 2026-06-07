package logging

import (
	"context"
)

var std = NewLogger()

func Init(opts ...Option) {
	std = NewLogger(opts...)
}

func L() *Logger {
	return std
}

func Debug(ctx context.Context, msg string) {
	std.Debug(ctx).Caller(1).Msg(msg)
}
func Info(ctx context.Context, msg string) {
	std.Info(ctx).Caller(1).Msg(msg)
}
func Warn(ctx context.Context, msg string) {
	std.Warn(ctx).Caller(1).Msg(msg)
}
func Error(ctx context.Context, msg string) {
	std.Error(ctx).Caller(1).Msg(msg)
}
func Errorw(ctx context.Context, err error) {
	std.Error(ctx).Caller(1).Err(err).Msg("error")
}
