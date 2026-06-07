package logging

import (
	"context"
)

type Logger struct {
	base       logger
	callerSkip int // 额外 skip 层数
}

func NewLogger(opts ...Option) *Logger {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return &Logger{
		base: newZerolog(cfg),
	}
}

func (l Logger) clone() Logger {
	return l
}

// With 返回新的 Logger
func (l Logger) With() Logger {
	return l.clone()
}

// WithCaller 启用 caller，并设置额外 skip
func (l Logger) WithCaller(skip int) Logger {
	nl := l.clone()
	nl.callerSkip = skip
	nl.base = nl.base.With().
		CallerWithSkipFrameCount(skip + 1). // +1 因为本函数占一帧
		Logger()
	return nl
}

// attachFields 应用 ctx 中的 Field 到 Event
func (l Logger) attachFields(ctx context.Context, e *Event) *Event {
	if e == nil {
		return nil
	}

	// 从 ctx 获取 Field
	if ctx != nil {
		for _, f := range FieldsFrom(ctx) {
			f(e)
		}
	}

	return e
}

func (l Logger) Debug(ctx context.Context) *Event {
	return l.attachFields(ctx, l.base.Debug())
}

func (l Logger) Info(ctx context.Context) *Event {
	return l.attachFields(ctx, l.base.Info())
}

func (l Logger) Warn(ctx context.Context) *Event {
	return l.attachFields(ctx, l.base.Warn())
}

func (l Logger) Error(ctx context.Context) *Event {
	return l.attachFields(ctx, l.base.Error())
}
