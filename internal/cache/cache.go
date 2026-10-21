// Package cache provides the engine's persistent artifact store.
//
// It exists to keep two concerns apart that the pipeline used to conflate.
// A pipeline run is handed an *execution workspace*: a scratch directory the
// caller owns, may hand over empty and may delete the moment the run ends.
// Expensive stage outputs — ASR transcripts, LLM extractions — must outlive
// that directory, so they are written here instead, to a location the engine
// owns and the caller never names.
//
// The store is a dumb, size-bounded blob store: keys are opaque strings
// chosen by the caller (the pipeline derives them from an episode's enclosure
// URL) and values are bytes. It knows nothing about episodes, slugs or output
// filenames, which is what allows an object-store implementation to be
// substituted later without touching the pipeline.
//
// # Eviction
//
// The store is bounded by a total byte budget (see NewStore). When a write
// pushes the directory over that budget, entries are deleted least-recently-
// used first until it fits again; reads refresh an entry's recency. A value
// larger than the whole budget is refused rather than stored and immediately
// evicted. Eviction is best-effort: a failure to delete is logged by the
// caller, never fatal, because the cache is an optimisation and losing it can
// only cost time.
package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// DefaultMaxBytes bounds a store created without an explicit budget. A
// transcript of an hour-long episode is tens of kilobytes, so this holds
// thousands of episodes while staying small enough for a disk-constrained
// host.
const DefaultMaxBytes int64 = 256 << 20 // 256 MiB

// tmpPrefix marks a partially written entry. Files carrying it are not
// readable entries and are excluded from the size accounting.
const tmpPrefix = ".tmp-"

// tmpMaxAge is how long an orphaned partial write (from a process killed
// mid-Put) is left alone before eviction reclaims it.
const tmpMaxAge = time.Hour

// ErrKeyInvalid is returned for a key that cannot safely name a file.
var ErrKeyInvalid = errors.New("cache: invalid key")

// Store is a size-bounded, LRU-evicting blob store backed by a directory.
// The zero value is not usable; call NewStore. A Store is safe for concurrent
// use: the pipeline processes episodes in parallel.
type Store struct {
	dir      string
	maxBytes int64

	mu sync.Mutex
}

// NewStore opens (creating if needed) a blob store rooted at dir and bounded
// to maxBytes total. A non-positive maxBytes selects DefaultMaxBytes.
//
// dir must be a location the engine owns and that persists across runs — it is
// never the pipeline's execution workspace.
func NewStore(dir string, maxBytes int64) (*Store, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("cache: directory is required")
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cache: create %s: %w", dir, err)
	}
	return &Store{dir: dir, maxBytes: maxBytes}, nil
}

// Dir reports where the store keeps its entries, for logging.
func (s *Store) Dir() string { return s.dir }

// MaxBytes reports the store's total size budget.
func (s *Store) MaxBytes() int64 { return s.maxBytes }

// Get returns the value stored under key. The boolean reports a hit; a miss,
// an unreadable entry and an invalid key are all reported as a miss, because
// a caller can only ever respond by recomputing the value.
func (s *Store) Get(key string) ([]byte, bool) {
	path, err := s.path(key)
	if err != nil {
		return nil, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	// Mark the entry as recently used. Eviction orders by modification time,
	// so a read that fails to touch only risks evicting a live entry early.
	now := time.Now()
	_ = os.Chtimes(path, now, now)

	return data, true
}

// Put stores data under key, replacing any existing value, then enforces the
// size budget. A value larger than the budget is refused.
func (s *Store) Put(key string, data []byte) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if int64(len(data)) > s.maxBytes {
		return fmt.Errorf("cache: value for %s is %d bytes, budget is %d", key, len(data), s.maxBytes)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tmp, err := os.CreateTemp(s.dir, tmpPrefix+"*")
	if err != nil {
		return fmt.Errorf("cache: create temp in %s: %w", s.dir, err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once renamed

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("cache: write %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cache: close %s: %w", tmpName, err)
	}
	// Rename is atomic within the directory, so a concurrent reader sees
	// either the old value or the complete new one, never a partial write.
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("cache: commit %s: %w", key, err)
	}

	return s.evict()
}

// entry is one stored blob, for eviction accounting.
type entry struct {
	path string
	size int64
	used time.Time
}

// evict deletes least-recently-used entries until the directory fits the
// budget. The caller must hold s.mu.
func (s *Store) evict() error {
	names, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("cache: read %s: %w", s.dir, err)
	}

	var (
		total   int64
		entries []entry
		errs    []error
		cutoff  = time.Now().Add(-tmpMaxAge)
	)
	for _, d := range names {
		if d.IsDir() {
			continue
		}
		info, err := d.Info()
		if err != nil {
			continue // vanished under us; nothing to account for
		}
		if strings.HasPrefix(d.Name(), tmpPrefix) {
			// An in-flight write belongs to whoever created it. Only reclaim
			// one old enough that its writer cannot still be running.
			if info.ModTime().Before(cutoff) {
				if err := os.Remove(filepath.Join(s.dir, d.Name())); err != nil && !os.IsNotExist(err) {
					errs = append(errs, err)
				}
			}
			continue
		}
		total += info.Size()
		entries = append(entries, entry{
			path: filepath.Join(s.dir, d.Name()),
			size: info.Size(),
			used: info.ModTime(),
		})
	}

	if total <= s.maxBytes {
		return errors.Join(errs...)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].used.Before(entries[j].used) })
	for _, e := range entries {
		if total <= s.maxBytes {
			break
		}
		if err := os.Remove(e.path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
			continue
		}
		total -= e.size
	}

	return errors.Join(errs...)
}

// path maps a key to its file, rejecting anything that could escape the
// store's directory or collide with the temp-file namespace.
func (s *Store) path(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("%w: empty", ErrKeyInvalid)
	}
	if len(key) > 200 {
		return "", fmt.Errorf("%w: %d characters exceeds 200", ErrKeyInvalid, len(key))
	}
	if strings.HasPrefix(key, ".") {
		return "", fmt.Errorf("%w: %q starts with a dot", ErrKeyInvalid, key)
	}
	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.':
		default:
			return "", fmt.Errorf("%w: %q contains %q", ErrKeyInvalid, key, r)
		}
	}
	return filepath.Join(s.dir, key), nil
}
