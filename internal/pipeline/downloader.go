package pipeline

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/afero"
)

type Downloader struct{}

func NewDownloader() *Downloader {
	return &Downloader{}
}

// DownloadAudio downloads an audio file concurrently and robustly using a streaming HTTP GET request,
// writing the result into the provided afero.Fs workspace.
func (d *Downloader) DownloadAudio(ctx context.Context, fs afero.Fs, url string, dest string) (err error) {
	if err := fs.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory in workspace: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		cerr := resp.Body.Close()
		if err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	dstFile, err := fs.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open destination file in workspace: %w", err)
	}

	defer func() {
		cerr := dstFile.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err := io.Copy(dstFile, resp.Body); err != nil {
		return fmt.Errorf("failed to stream download to workspace: %w", err)
	}

	return nil
}

// FetchMetadata fetches the file size and checks for Range request support using an HTTP HEAD request.
func (d *Downloader) FetchMetadata(url string) (size int64, supportsRange bool, err error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, false, fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer func() {
		cerr := resp.Body.Close()
		if err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, false, fmt.Errorf("missing Content-Length header")
	}

	size, err = strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse Content-Length: %w", err)
	}

	supportsRange = resp.Header.Get("Accept-Ranges") == "bytes"

	return size, supportsRange, nil
}
