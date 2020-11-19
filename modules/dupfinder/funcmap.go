package dupfinder

import (
	"strings"
	"text/template"
)

var (
	escaper = strings.NewReplacer(
		"+", "\\+", "-", "\\-", "=", "\\=", "&", "\\&",
		"|", "\\|", ">", "\\>", "<", "\\<", "!", "\\!",
		"(", "\\(", ")", "\\)", "{", "\\{", "}", "\\}",
		"[", "\\[", "]", "\\]", "^", "\\^", "\\", "\\\\",
		"\"", "\\\"", "~", "\\~", "*", "\\*", "?", "\\?",
		":", "\\:", "/", "\\/", " ", "\\ ",
	)

	funcmap = template.FuncMap{
		"escape": escapeQuery,
	}
)

// Escape char from https://blevesearch.com/docs/Query-String-Query/
func escapeQuery(q string) string {
	return escaper.Replace(q)
}
