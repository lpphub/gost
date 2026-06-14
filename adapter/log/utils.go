package log

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + " ...[truncated]"
	}
	return s
}
