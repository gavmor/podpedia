package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/podpedia/pkg/naming"
	"github.com/spf13/afero"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) SaveRawData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := fs.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Printf("[Storage] Saving raw data for %s\n", ep.ID)
	fileName := filepath.Join(outputDir, naming.RawName(ep.ID))
	return afero.WriteFile(fs, fileName, []byte(ep.Transcript), 0644)
}

func (s *Store) SaveStructuredData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode, entry json.RawMessage, schemeID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := fs.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Printf("[Storage] Saving structured data for %s (scheme: %s)\n", ep.ID, schemeID)
	fileName := filepath.Join(outputDir, naming.StructuredName(ep.ID, schemeID))
	return afero.WriteFile(fs, fileName, entry, 0644)
}
