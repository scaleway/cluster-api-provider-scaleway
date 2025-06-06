package scope

// truncateString shortens a string to a maximum of maxLen characters.
// If the string exceeds this length, it replaces the middle portion
// with a single dash, preserving characters from the start and end.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	n := (maxLen - 1) / 2
	return s[:n] + "-" + s[len(s)-(maxLen-1-n):]
}
