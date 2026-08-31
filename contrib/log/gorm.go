package log

import (
	"context"
	"errors"
	"fmt"
	"time"

	glog "github.com/lpphub/gost/log"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gmlog "gorm.io/gorm/logger"
)

type GORMLogCfg struct {
	SlowThreshold time.Duration
	SQLMaxLen     int
	CallerSkip    int
	Level         zerolog.Level
}

type GORMLogger struct {
	cfg GORMLogCfg
}

func NewGORMLogger(cfg GORMLogCfg) gmlog.Interface {
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = 2000 * time.Millisecond
	}
	if cfg.SQLMaxLen == 0 {
		cfg.SQLMaxLen = 1024
	}
	return &GORMLogger{cfg: cfg}
}

func (l *GORMLogger) LogMode(level gmlog.LogLevel) gmlog.Interface {
	cfg := l.cfg
	switch level {
	case gmlog.Silent:
		cfg.Level = zerolog.Disabled
	case gmlog.Error:
		cfg.Level = zerolog.ErrorLevel
	case gmlog.Warn:
		cfg.Level = zerolog.WarnLevel
	case gmlog.Info:
		cfg.Level = zerolog.InfoLevel
	}
	return &GORMLogger{cfg: cfg}
}

func (l *GORMLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	ev := l.event(ctx, zerolog.InfoLevel)
	if ev == nil {
		return
	}
	ev.Msg(fmt.Sprintf(msg, data...))
}

func (l *GORMLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	ev := l.event(ctx, zerolog.WarnLevel)
	if ev == nil {
		return
	}
	ev.Msg(fmt.Sprintf(msg, data...))
}

func (l *GORMLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	ev := l.event(ctx, zerolog.ErrorLevel)
	if ev == nil {
		return
	}
	ev.Msg(fmt.Sprintf(msg, data...))
}

func (l *GORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	sql = truncate(sql, l.cfg.SQLMaxLen)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		if ev := l.event(ctx, zerolog.ErrorLevel); ev != nil {
			ev.Err(fmt.Errorf("query error: %w", err)).
				Str("sql", sql).Int64("rows", rows).Int64("cost", elapsed.Milliseconds()).Msg("error")
		}
	case l.cfg.SlowThreshold > 0 && elapsed > l.cfg.SlowThreshold:
		if ev := l.event(ctx, zerolog.WarnLevel); ev != nil {
			ev.Str("sql", sql).Int64("rows", rows).Int64("cost", elapsed.Milliseconds()).Msg("slow query")
		}
	default:
		if ev := l.event(ctx, zerolog.InfoLevel); ev != nil {
			ev.Str("sql", sql).Int64("rows", rows).Int64("cost", elapsed.Milliseconds()).Msg("sql done")
		}
	}
}

func (l *GORMLogger) event(ctx context.Context, level zerolog.Level) *zerolog.Event {
	if l.cfg.Level != 0 && level < l.cfg.Level {
		return nil
	}
	var ev *zerolog.Event
	switch level {
	case zerolog.InfoLevel:
		ev = glog.Ctx(ctx).Info()
	case zerolog.WarnLevel:
		ev = glog.Ctx(ctx).Warn()
	case zerolog.ErrorLevel:
		ev = glog.Ctx(ctx).Error()
	default:
		return nil
	}
	return withCaller(ev, l.cfg.CallerSkip)
}
