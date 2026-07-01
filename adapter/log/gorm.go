package log

import (
	"context"
	"errors"
	"fmt"
	"time"

	glog "github.com/lpphub/gost/log"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const defaultGormCallerSkip = 5

type GormLogger struct {
	logLevel      gormlogger.LogLevel
	slowThreshold time.Duration
	sqlMaxLen     int
	callerSkip    int
}

type GormLoggerOption func(*GormLogger)

func WithCallerSkip(skip int) GormLoggerOption {
	return func(l *GormLogger) {
		l.callerSkip = skip
	}
}

func NewGormLogger(opts ...GormLoggerOption) gormlogger.Interface {
	l := &GormLogger{
		logLevel:      gormlogger.Info,
		slowThreshold: 1000 * time.Millisecond,
		sqlMaxLen:     1024,
		callerSkip:    defaultGormCallerSkip,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	cp := *l
	cp.logLevel = level
	return &cp
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logf(ctx, gormlogger.Info, msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logf(ctx, gormlogger.Warn, msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logf(ctx, gormlogger.Error, msg, data...)
}

func (l *GormLogger) logf(ctx context.Context, level gormlogger.LogLevel, msg string, data ...interface{}) {
	if l.logLevel < level {
		return
	}
	if len(data) > 0 {
		msg = fmt.Sprintf(msg, data...)
	}

	switch level {
	case gormlogger.Warn:
		glog.WarnwDepth(ctx, l.callerSkip, msg)
	case gormlogger.Error:
		glog.ErrorwDepth(ctx, l.callerSkip, msg)
	default:
		glog.InfowDepth(ctx, l.callerSkip, msg)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	sql = truncate(sql, l.sqlMaxLen)

	fields := []glog.Field{
		glog.Str("sql", sql),
		glog.Int64("rows", rows),
		glog.Int64("cost", elapsed.Milliseconds()),
	}

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		glog.ErrorwfDepth(ctx, l.callerSkip, fmt.Errorf("query error: %w", err), fields...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		glog.WarnwDepth(ctx, l.callerSkip, "slow query", fields...)
	default:
		glog.InfowDepth(ctx, l.callerSkip, "sql done", fields...)
	}
}
