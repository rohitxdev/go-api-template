package util

import "strings"

// Returns a reversed string
func ReverseString(s string) string {
	var sb strings.Builder
	runes := []rune(s)
	for i := len(runes) - 1; 0 <= i; i-- {
		sb.WriteRune(runes[i])
	}
	return sb.String()
}
