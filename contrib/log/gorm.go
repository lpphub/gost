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

type GormLogCfg struct {
	SlowThreshold time.Duration
	SQLMaxLen     int
	CallerSkip    int
	Level         zerolog.Level // 0 = inherit global level; non-zero = override
}

type GormLogger struct {
	cfg GormLogCfg
}

func NewGormLogger(cfg GormLogCfg) gmlog.Interface {
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = 2000 * time.Millisecond
	}
	if cfg.SQLMaxLen == 0 {
		cfg.SQLMaxLen = 1024
	}
	return &GormLogger{cfg: cfg}
}

func (l *GormLogger) LogMode(level gmlog.LogLevel) gmlog.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	ev := l.event(ctx, zerolog.InfoLevel)
	if ev == nil {
		return
	}
	ev.Msg(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	ev := l.event(ctx, zerolog.WarnLevel)
	if ev == nil {
		return
	}
	ev.Msg(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	ev := l.event(ctx, zerolog.ErrorLevel)
	if ev == nil {
		return
	}
	ev.Msg(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
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

// event returns a *zerolog.Event at the given level with caller attached,
// or nil when the level is below the configured threshold.
func (l *GormLogger) event(ctx context.Context, level zerolog.Level) *zerolog.Event {
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
