package logx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lpphub/gost/logger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	logLevel      gormlogger.LogLevel
	slowThreshold time.Duration
	sqlMaxLen     int
}

func NewGormLogger() gormlogger.Interface {
	return &GormLogger{
		logLevel:      gormlogger.Info,
		slowThreshold: 1000 * time.Millisecond,
		sqlMaxLen:     1024,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l.logLevel = level
	return l
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
		logger.Warnf(ctx, msg)
	case gormlogger.Error:
		logger.Errorf(ctx, msg)
	default:
		logger.Infof(ctx, msg)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if len(sql) > l.sqlMaxLen {
		sql = sql[:l.sqlMaxLen] + " ...[truncated]"
	}

	fields := []logger.Field{
		logger.Str("sql", sql),
		logger.Int64("rows", rows),
		logger.Dur("duration", elapsed),
	}

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		logger.Errorwf(ctx, fmt.Errorf("query error: %w", err), fields...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		logger.Warnf(ctx, "slow query", fields...)
	default:
		logger.Infof(ctx, "query success", fields...)
	}
}

