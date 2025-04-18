package crawler

import (
	"net/http"
	"time"
)

func CheckWebsite(url string) (int, int64) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0
	}
	defer resp.Body.Close()

	responseTime := time.Since(start).Milliseconds()
	return resp.StatusCode, responseTime
}
