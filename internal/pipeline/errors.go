package pipeline

import "errors"

var (
	ErrTranscriptionTimeout = errors.New("transcription timeout")
	ErrInvalidFeed          = errors.New("invalid RSS feed")
	ErrDownloadFailed       = errors.New("download failed")
	ErrExtractionFailed     = errors.New("extraction failed")
	ErrStoreFailed          = errors.New("storage failed")
)
