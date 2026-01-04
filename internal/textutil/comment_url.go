package textutil

import (
	"regexp"
	"strings"
)

func ExtractURLFromComment(comment string) string {
	if comment == "" {
		return ""
	}

	urlPattern := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches := urlPattern.FindString(comment)
	if matches != "" {
		return matches
	}

	rutrackerPattern := regexp.MustCompile(`(?:rutracker\.org|rutracker\.cc)/[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches = rutrackerPattern.FindString(comment)
	if matches != "" {
		if !strings.HasPrefix(matches, "http://") && !strings.HasPrefix(matches, "https://") {
			return "https://" + matches
		}

		return matches
	}

	return ""
}
