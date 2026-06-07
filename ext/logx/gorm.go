package logx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lpphub/goweb/pkg/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormLogger 自定义GORM日志记录器
type GormLogger struct {
	logger        logging.Logger
	logLevel      logger.LogLevel
	slowThreshold time.Duration
	sqlMaxLen     int
}

// NewGormLogger 创建新的GORM日志记录器
func NewGormLogger() logger.Interface {
	return &GormLogger{
		logLevel:      logger.Info,
		logger:        logging.L().WithCaller(5),
		slowThreshold: 1000 * time.Millisecond,
		sqlMaxLen:     1024,
	}
}

// LogMode 设置日志等级
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.logLevel = level
	return l
}

// Info 记录信息日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		if len(data) > 0 {
			msg = fmt.Sprintf(msg, data...)
		}
		l.log(ctx, logger.Info, msg, nil)
	}
}

// Warn 记录警告日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		if len(data) > 0 {
			msg = fmt.Sprintf(msg, data...)
		}
		l.log(ctx, logger.Warn, msg, nil)
	}
}

// Error 记录错误日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		if len(data) > 0 {
			msg = fmt.Sprintf(msg, data...)
		}
		l.log(ctx, logger.Error, msg, nil)
	}
}

// Trace 记录 SQL 执行追踪
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 截断超长 SQL
	if len(sql) > l.sqlMaxLen {
		sql = sql[:l.sqlMaxLen] + " ...[truncated]"
	}

	fields := map[string]interface{}{
		"sql":         sql,
		"rows":        rows,
		"duration_ms": elapsed.Milliseconds(),
	}

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.log(ctx, logger.Error, fmt.Sprintf("query error: %s", err.Error()), fields)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		l.log(ctx, logger.Warn, "slow query", fields)
	default:
		l.log(ctx, logger.Info, "query success", fields)
	}
}

// log 通用日志方法
func (l *GormLogger) log(ctx context.Context, level logger.LogLevel, msg string, fields map[string]interface{}) {
	if level < l.logLevel {
		return
	}

	switch level {
	case logger.Warn:
		l.logger.Warn(ctx).Fields(fields).Msg(msg)
	case logger.Error:
		l.logger.Error(ctx).Fields(fields).Msg(msg)
	default:
		l.logger.Info(ctx).Fields(fields).Msg(msg)
	}
}
