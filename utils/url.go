package utils

import (
	"regexp"
	"strings"
)

// NormalizeURL formats the target URL
func NormalizeURL(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	re := regexp.MustCompile(`^(?:https?://)?(?:www\.)?([^:/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return url
}