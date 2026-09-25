package utils

import "unicode/utf8"

// RuneLength returns the number of runes (Unicode characters) in s.
// Unlike len(s) which counts bytes, this counts characters — a CJK
// character (3 bytes in UTF-8) counts as 1.
func RuneLength(s string) int {
	return utf8.RuneCountInString(s)
}
