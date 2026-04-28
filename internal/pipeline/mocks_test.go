package pipeline_test

import (
	"context"
	"fmt"
	"path/filepath"

	. "github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/spf13/afero"
)

// Mock implementations for testing
type mockTranscriber struct{}

func (m *mockTranscriber) Transcribe(ctx context.Context, audioURL string, prompt string) (string, error) {
	return "mock transcript", nil
}

type mockExtractor struct{}

func (m *mockExtractor) ExtractEntities(ctx context.Context, ep *types.Episode, scheme []byte) ([]byte, error) {
	return []byte(`{"episode_id": "` + ep.ID + `", "mocked": true}`), nil
}

type mockDownloader struct {
	fs afero.Fs
}

func (m *mockDownloader) DownloadAudio(ctx context.Context, fs afero.Fs, url string, dest string) error {
	if fs == nil {
		fs = m.fs
	}
	if fs == nil {
		fs = afero.NewOsFs()
	}
	if err := fs.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return afero.WriteFile(fs, dest, []byte("mock audio"), 0644)
}

type mockStore struct {
	fs              afero.Fs
	RawSaved        bool
	StructuredSaved bool
	savedStructured map[string][]byte
}

func (m *mockStore) SaveRawData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode) error {
	m.RawSaved = true
	return nil
}

func (m *mockStore) SaveStructuredData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode, entry []byte, schemeID string) error {
	m.StructuredSaved = true

	id := ep.ID
	if id == "" {
		id = "unknown"
	}

	// Using the real slug helper from the pipeline package via dot-import
	key := fmt.Sprintf("%s_%s", Slug(id), schemeID)
	m.savedStructured[key] = entry
	return nil
}
