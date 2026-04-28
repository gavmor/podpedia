package storage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
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
	// Sanitize ep.ID using pipeline.Slug
	fileName := filepath.Join(outputDir, pipeline.Slug(ep.ID)+"_raw.txt")
	return afero.WriteFile(fs, fileName, []byte(ep.Transcript), 0644)
}

func (s *Store) SaveStructuredData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode, entry []byte, schemeID string) error {
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
	// Sanitize ep.ID using pipeline.Slug
	fileName := filepath.Join(outputDir, fmt.Sprintf("%s_%s.json", pipeline.Slug(ep.ID), schemeID))
	return afero.WriteFile(fs, fileName, entry, 0644)
}
