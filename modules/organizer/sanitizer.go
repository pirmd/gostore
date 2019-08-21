package organizer

import (
	"strings"
	"unicode"
)

// pathSanitizer rewrites string to remove non-standard path characters
func pathSanitizer(path string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) ||
			unicode.IsDigit(r) ||
			unicode.IsMark(r) ||
			r == '.' ||
			r == '/' ||
			r == '\\' ||
			r == '_' ||
			r == '-' ||
			r == '%' ||
			r == ' ' ||
			r == '#' {
			return r
		}

		if unicode.IsSpace(r) {
			return ' '
		}

		if unicode.In(r, unicode.Hyphen) {
			return '-'
		}

		return -1
	}, path)
}

// nospaceSanitizer rewrites string to remove unreasonable path characters
func nospaceSanitizer(path string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) ||
			unicode.IsDigit(r) ||
			unicode.IsMark(r) ||
			r == '.' ||
			r == '/' ||
			r == '\\' ||
			r == '_' ||
			r == '-' ||
			r == '%' ||
			r == '#' {
			return r
		}

		if unicode.IsSpace(r) {
			return '_'
		}

		if unicode.In(r, unicode.Hyphen) ||
			r == '\'' {
			return '-'
		}

		return -1
	}, path)
}
