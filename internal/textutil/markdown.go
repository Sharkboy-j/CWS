package textutil

import "strings"

// EscapeMarkdown escapes a minimal set of characters for Telegram legacy Markdown parse mode.
func EscapeMarkdown(s string) string {
	r := strings.NewReplacer(
		`\\`, `\\\\`,
		`_`, `\_`,
		`*`, `\*`,
		"`", "\\`",
		`[`, `\[`,
	)

	return r.Replace(s)
}
