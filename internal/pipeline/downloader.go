package pipeline

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cavaliergopher/grab/v3"
)

type Downloader struct{}

func NewDownloader() *Downloader {
	return &Downloader{}
}

// DownloadAudio downloads an audio file concurrently and robustly using the grab library.
func (d *Downloader) DownloadAudio(url string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	_, err := grab.Get(dest, url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// FetchMetadata fetches the file size and checks for Range request support using an HTTP HEAD request.
func (d *Downloader) FetchMetadata(url string) (int64, bool, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, false, fmt.Errorf("failed to execute HEAD request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
