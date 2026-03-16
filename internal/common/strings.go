package common

import "strings"

// Split2 splits s by the first occurrence of sep. If sep is not present or
// either side is empty, it returns nil.
func Split2(s, sep string) []string {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) != 2 {
		return nil
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return nil
	}
	return parts
}
