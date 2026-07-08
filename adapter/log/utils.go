package log

import "github.com/rs/zerolog"

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + " ...[truncated]"
	}
	return s
}

func addCaller(ev *zerolog.Event, skip int) *zerolog.Event {
	if skip > 0 {
		return ev.Caller(skip)
	}
	return ev
}
