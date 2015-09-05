package env

import (
	"strings"
	"unicode"
)

// IdentityMapper returns the input string
func IdentityMapper(str string) string {
	return str
}

// UnderscoreMapper converts CamelCase strings to their camel_case
// counterpart
func UnderscoreMapper(str string) string {
	var (
		parts = []string{}
		cur   = []rune{}
		last2 = [2]rune{}
	)

	for _, c := range str {
		if unicode.IsUpper(c) {
			if last2[1] != 0 && unicode.IsLower(last2[1]) {
				parts = append(parts, string(cur))
				cur = nil
			}

			cur = append(cur, unicode.ToLower(c))
		} else {
			if last2[0] != 0 && last2[1] != 0 && unicode.IsUpper(last2[0]) && unicode.IsUpper(last2[1]) {
				parts = append(parts, string(cur[:len(cur)-1]))
				cur = []rune{cur[len(cur)-1]}
			}

			cur = append(cur, c)
		}

		last2[0] = last2[1]
		last2[1] = c
	}

	if len(cur) > 0 {
		parts = append(parts, string(cur))
	}

	return strings.Join(parts, "_")
}
