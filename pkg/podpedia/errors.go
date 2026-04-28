package podpedia

import "github.com/gavmor/podpedia/internal/pipeline"

var (
	ErrTranscriptionTimeout = pipeline.ErrTranscriptionTimeout
	ErrInvalidFeed          = pipeline.ErrInvalidFeed
	ErrDownloadFailed       = pipeline.ErrDownloadFailed
	ErrExtractionFailed     = pipeline.ErrExtractionFailed
	ErrStoreFailed          = pipeline.ErrStoreFailed
)
