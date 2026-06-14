package log

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	glog "github.com/lpphub/gost/log"
	"github.com/redis/go-redis/v9"
)

type RedisLogger struct {
	slowThreshold time.Duration
	cmdMaxLen     int
}

func NewRedisLogger() *RedisLogger {
	return &RedisLogger{
		slowThreshold: 100 * time.Millisecond,
		cmdMaxLen:     1024,
	}
}

func (l *RedisLogger) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()
		conn, err := next(ctx, network, addr)

		fields := []glog.Field{
			glog.Str("addr", addr),
			glog.Dur("duration", time.Since(start)),
		}

		if err != nil {
			glog.Errorf(ctx, "redis connected failed", append(fields, glog.Err(err))...)
		} else {
			glog.Infof(ctx, "redis connected", fields...)
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
			glog.Dur("duration", elapsed),
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
			glog.Dur("duration", elapsed),
		}

		l.logResult(ctx, fields, err, elapsed, "redis pipeline success", "redis pipeline slow", "redis pipeline error")
		return err
	}
}

func (l *RedisLogger) logResult(ctx context.Context, fields []glog.Field, err error, elapsed time.Duration, okMsg, slowMsg, errMsg string) {
	switch {
	case err != nil && !errors.Is(err, redis.Nil):
		glog.Errorf(ctx, errMsg, append(fields, glog.Err(err))...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		glog.Warnf(ctx, slowMsg, fields...)
	default:
		glog.Infof(ctx, okMsg, fields...)
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
		fmt.Fprintf(&sb, "%v", arg)
	}
	s := sb.String()
	if len(s) > l.cmdMaxLen {
		s = s[:l.cmdMaxLen] + " ...[truncated]"
	}
	return s
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
			sb.WriteString(fmt.Sprintf("... (%d more)", len(cmds)-5))
			break
		}
		sb.WriteString(l.buildCmd(cmd))
	}
	s := sb.String()
	if len(s) > l.cmdMaxLen {
		s = s[:l.cmdMaxLen] + " ...[truncated]"
	}
	return s
}
