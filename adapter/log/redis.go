package log

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	glog "github.com/lpphub/gost/log"
	"github.com/redis/go-redis/v9"
)

const defaultRedisCallerSkip = 4

type RedisLogger struct {
	slowThreshold time.Duration
	cmdMaxLen     int
	callerSkip    int
}

type RedisLoggerOption func(*RedisLogger)

func WithRedisCallerSkip(skip int) RedisLoggerOption {
	return func(l *RedisLogger) {
		l.callerSkip = skip
	}
}

func NewRedisLogger(opts ...RedisLoggerOption) *RedisLogger {
	l := &RedisLogger{
		slowThreshold: 100 * time.Millisecond,
		cmdMaxLen:     1024,
		callerSkip:    defaultRedisCallerSkip,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *RedisLogger) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()
		conn, err := next(ctx, network, addr)

		fields := []glog.Field{
			glog.Str("addr", addr),
			glog.Int64("cost", time.Since(start).Milliseconds()),
		}

		if err != nil {
			glog.ErrorwDepth(ctx, l.callerSkip, "redis connected failed", append(fields, glog.Err(err))...)
		} else {
			glog.InfowDepth(ctx, l.callerSkip, "redis connected", fields...)
		}
		return conn, err
	}
}

func (l *RedisLogger) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		elapsed := time.Since(start)

		fields := []glog.Field{
			glog.Str("cmd", l.buildCmd(cmd)),
			glog.Int64("cost", elapsed.Milliseconds()),
		}

		l.logResult(ctx, fields, err, elapsed, "redis success", "redis slow", "redis error")
		return err
	}
}

func (l *RedisLogger) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		elapsed := time.Since(start)

		fields := []glog.Field{
			glog.Str("cmd", l.buildPipelineCmd(cmds)),
			glog.Int64("cost", elapsed.Milliseconds()),
		}

		l.logResult(ctx, fields, err, elapsed, "redis pipeline success", "redis pipeline slow", "redis pipeline error")
		return err
	}
}

func (l *RedisLogger) logResult(ctx context.Context, fields []glog.Field, err error, elapsed time.Duration, okMsg, slowMsg, errMsg string) {
	switch {
	case err != nil && !errors.Is(err, redis.Nil):
		glog.ErrorwDepth(ctx, l.callerSkip, errMsg, append(fields, glog.Err(err))...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		glog.WarnwDepth(ctx, l.callerSkip, slowMsg, fields...)
	default:
		glog.InfowDepth(ctx, l.callerSkip, okMsg, fields...)
	}
}

func (l *RedisLogger) buildCmd(cmd redis.Cmder) string {
	args := cmd.Args()
	if len(args) == 0 {
		return cmd.Name()
	}
	var sb strings.Builder
	for i, arg := range args {
		if i > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprint(&sb, arg)
	}
	return truncate(sb.String(), l.cmdMaxLen)
}

func (l *RedisLogger) buildPipelineCmd(cmds []redis.Cmder) string {
	if len(cmds) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, cmd := range cmds {
		if i > 0 {
			sb.WriteString("; ")
		}
		if i >= 5 {
			sb.WriteString("... ("); sb.WriteString(strconv.Itoa(len(cmds)-5)); sb.WriteString(" more)")
			break
		}
		sb.WriteString(l.buildCmd(cmd))
	}
	return truncate(sb.String(), l.cmdMaxLen)
}
