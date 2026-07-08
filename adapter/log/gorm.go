package log

import (
	"context"
	"errors"
	"fmt"
	"time"

	glog "github.com/lpphub/gost/log"
	"gorm.io/gorm"
	gmlog "gorm.io/gorm/logger"
)

type GormLogCfg struct {
	SlowThreshold time.Duration
	SQLMaxLen     int
	CallerSkip    int
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
	if len(data) > 0 {
		msg = fmt.Sprintf(msg, data...)
	}
	addCaller(glog.Ctx(ctx).Info(), l.cfg.CallerSkip).Msg(msg)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if len(data) > 0 {
		msg = fmt.Sprintf(msg, data...)
	}
	addCaller(glog.Ctx(ctx).Warn(), l.cfg.CallerSkip).Msg(msg)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if len(data) > 0 {
		msg = fmt.Sprintf(msg, data...)
	}
	addCaller(glog.Ctx(ctx).Error(), l.cfg.CallerSkip).Msg(msg)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	sql = truncate(sql, l.cfg.SQLMaxLen)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		addCaller(glog.Ctx(ctx).Error().Err(fmt.Errorf("query error: %w", err)).
			Str("sql", sql).Int64("rows", rows).Int64("cost", elapsed.Milliseconds()),
			l.cfg.CallerSkip).Msg("error")
	case l.cfg.SlowThreshold > 0 && elapsed > l.cfg.SlowThreshold:
		addCaller(glog.Ctx(ctx).Warn().
			Str("sql", sql).Int64("rows", rows).Int64("cost", elapsed.Milliseconds()),
			l.cfg.CallerSkip).Msg("slow query")
	default:
		addCaller(glog.Ctx(ctx).Info().
			Str("sql", sql).Int64("rows", rows).Int64("cost", elapsed.Milliseconds()),
			l.cfg.CallerSkip).Msg("sql done")
	}
}
