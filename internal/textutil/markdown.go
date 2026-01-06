package textutil

import "strings"

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
