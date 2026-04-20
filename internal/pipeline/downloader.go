package pipeline

import (
	"fmt"
	"net/http"
	"strconv"
)

// GetAudioMetadata fetches the file size and checks for Range request support using an HTTP HEAD request.
func GetAudioMetadata(url string) (int64, bool, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, false, fmt.Errorf("failed to execute HEAD request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, false, fmt.Errorf("missing Content-Length header")
	}

	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse Content-Length: %w", err)
	}

	supportsRange := resp.Header.Get("Accept-Ranges") == "bytes"

	return size, supportsRange, nil
}
