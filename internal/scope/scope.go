package scope

// maxResourceNameLen is the maximum length a Scaleway resource name should have.
const maxResourceNameLen = 128

// truncateString shortens a string to a maximum of 128 characters.
// If the string exceeds this length, it replaces the middle portion
// with a single dash, preserving characters from the start and end.
func truncateString(s string) string {
	if len(s) <= maxResourceNameLen {
		return s
	}
	n := (maxResourceNameLen - 1) / 2
	return s[:n] + "-" + s[len(s)-(maxResourceNameLen-1-n):]
}
