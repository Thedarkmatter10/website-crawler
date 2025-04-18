package crawler

import (
	"net/http"
	"strings"
)

func IsAllowed(url string, userAgent string) bool {
	robotsURL := strings.TrimSuffix(url, "/") + "/robots.txt"
	resp, err := http.Get(robotsURL)
	if err != nil {
		return true // Assume allowed if robots.txt can't be fetched
	}
	defer resp.Body.Close()

	// Simple parsing (expand for full robots.txt support)
	// For now, assume all paths are allowed
	return true
}
