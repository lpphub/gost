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

type RedisLogCfg struct {
	SlowThreshold time.Duration
	CmdMaxLen     int
	CallerSkip    int
}

type RedisLogger struct {
	cfg RedisLogCfg
}

func NewRedisLogger(cfg RedisLogCfg) *RedisLogger {
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = 100 * time.Millisecond
	}
	if cfg.CmdMaxLen == 0 {
		cfg.CmdMaxLen = 1024
	}
	return &RedisLogger{cfg: cfg}
}

func (l *RedisLogger) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()
		conn, err := next(ctx, network, addr)
		cost := time.Since(start).Milliseconds()

		if err != nil {
			glog.Ctx(ctx).Error().Err(err).
				Str("addr", addr).Int64("cost", cost).
				Caller(l.cfg.CallerSkip).Msg("redis connected failed")
		} else {
			glog.Ctx(ctx).Info().
				Str("addr", addr).Int64("cost", cost).
				Caller(l.cfg.CallerSkip).Msg("redis connected")
		}
		return conn, err
	}
}

func (l *RedisLogger) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		elapsed := time.Since(start)

		l.logResult(ctx, err, elapsed, l.buildCmd(cmd), "redis success", "redis slow", "redis error")
		return err
	}
}

func (l *RedisLogger) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		elapsed := time.Since(start)

		l.logResult(ctx, err, elapsed, l.buildPipelineCmd(cmds), "redis pipeline success", "redis pipeline slow", "redis pipeline error")
		return err
	}
}

func (l *RedisLogger) logResult(ctx context.Context, err error, elapsed time.Duration, cmd, okMsg, slowMsg, errMsg string) {
	switch {
	case err != nil && !errors.Is(err, redis.Nil):
		glog.Ctx(ctx).Error().Err(err).
			Str("cmd", cmd).Int64("cost", elapsed.Milliseconds()).
			Caller(l.cfg.CallerSkip).Msg(errMsg)
	case l.cfg.SlowThreshold > 0 && elapsed > l.cfg.SlowThreshold:
		glog.Ctx(ctx).Warn().
			Str("cmd", cmd).Int64("cost", elapsed.Milliseconds()).
			Caller(l.cfg.CallerSkip).Msg(slowMsg)
	default:
		glog.Ctx(ctx).Info().
			Str("cmd", cmd).Int64("cost", elapsed.Milliseconds()).
			Caller(l.cfg.CallerSkip).Msg(okMsg)
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
	return truncate(sb.String(), l.cfg.CmdMaxLen)
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
			sb.WriteString("... (")
			sb.WriteString(strconv.Itoa(len(cmds) - 5))
			sb.WriteString(" more)")
			break
		}
		sb.WriteString(l.buildCmd(cmd))
	}
	return truncate(sb.String(), l.cfg.CmdMaxLen)
}
