package kernel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/podpedia/pkg/naming"
)

// storeWasm is the filesystem-writing plugin this package's mount policy is
// exercised against. It is a build artifact, not a checked-in fixture:
// `make plugins-build` (and the CI job, which runs it before `go test`)
// produces it.
const storeWasm = "../../dist/plugins/store.wasm"

// TestAllowedPathsMapsEachGuestPathOnce is the policy half of the mount
// regression. The Extism SDK applies AllowedPaths by ranging over the Go map
// (extism/go-sdk plugin.go, `for host, guest := range
// p.manifest.AllowedPaths`), and Go randomises map iteration order, so two
// host directories that resolve to the SAME guest path are not a redundant
// pair — they are a coin flip, re-tossed on every Kernel.Call.
//
// The policy used to read {"/": "/", ".": "."}; wazero normalises the guest
// path "." to "/", so both entries claimed the guest root and the winner was
// drawn at random per call. Any entry added here has to resolve to a guest
// path no other entry claims.
func TestAllowedPathsMapsEachGuestPathOnce(t *testing.T) {
	seen := map[string]string{} // guest path -> host path that claimed it
	for host, guest := range AllowedPaths() {
		normalised := normaliseGuestPath(guest)
		if other, clash := seen[normalised]; clash {
			t.Errorf("hosts %q and %q both mount at guest path %q; "+
				"the SDK ranges over this map, so which one wins is random per call",
				other, host, normalised)
			continue
		}
		seen[normalised] = host
	}
}

// normaliseGuestPath reproduces the guest-side resolution wazero performs on a
// mount's guest path: a relative path is taken from the guest root, and the
// guest root is where a "." mount lands.
func normaliseGuestPath(guest string) string {
	if !strings.HasPrefix(guest, "/") {
		guest = "/" + guest
	}
	return filepath.Clean(guest)
}

// TestStoreWritesLandAtTheRequestedAbsolutePath is the behavioural half: it
// drives the real plugin through the real kernel config, from a working
// directory that is not "/", and insists every write lands where the host
// asked for it.
//
// With the ambiguous mount pair in place this failed by writing roughly half
// of its files to <cwd>/<output dir> instead — and the plugin reported
// success either way ({"path":"/tmp/..."}), so the only symptom a caller ever
// saw was missing output.
func TestStoreWritesLandAtTheRequestedAbsolutePath(t *testing.T) {
	wasm, err := os.ReadFile(storeWasm)
	if err != nil {
		t.Skipf("no %s; run `make plugins-build` first (CI does): %v", storeWasm, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if cwd == "/" {
		t.Fatal("this test only means anything from a working directory other than /")
	}

	outputDir := t.TempDir()
	shadowRoot := shadowRootFor(t, cwd, outputDir)

	pk := NewPodpediaKernel("", "", "")
	pk.Load("store", wasm)
	store := NewStore(pk)

	// One call is a coin toss; the defect only shows up over a run of them.
	const calls = 12
	var missing []string
	for i := 0; i < calls; i++ {
		id := fmt.Sprintf("mount-probe-%02d", i)
		ep := types.Episode{ID: id, Title: "Mount Probe", Transcript: "hi"}

		if err := store.SaveRawData(context.Background(), nil, outputDir, &ep); err != nil {
			t.Fatalf("call %d: SaveRawData reported an error: %v", i, err)
		}

		want := filepath.Join(outputDir, naming.RawName(id))
		if _, err := os.Stat(want); err != nil {
			missing = append(missing, want)
		}
	}

	if len(missing) > 0 {
		t.Errorf("%d of %d writes did not land at the requested absolute path; first missing: %s",
			len(missing), calls, missing[0])
	}

	if strayed := filesUnder(shadowRoot); len(strayed) > 0 {
		t.Errorf("%d write(s) landed under the working directory instead of the requested "+
			"absolute path (cwd-relative mount won): %v", len(strayed), strayed)
	}
}

// shadowRootFor returns the directory a cwd-relative mount would divert
// writes to — <cwd>/<first segment of abs> — and arranges for the diverted
// tree to be removed after the test, so a failing run does not leave debris
// in the checkout.
func shadowRootFor(t *testing.T, cwd, abs string) string {
	t.Helper()

	rel := strings.TrimPrefix(abs, "/")
	root := filepath.Join(cwd, strings.SplitN(rel, string(filepath.Separator), 2)[0])

	existedBefore := true
	if _, err := os.Stat(root); os.IsNotExist(err) {
		existedBefore = false
	}
	t.Cleanup(func() {
		if !existedBefore {
			_ = os.RemoveAll(root)
		}
	})

	return root
}

// filesUnder lists every regular file beneath root, which is empty when root
// does not exist.
func filesUnder(root string) []string {
	var found []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // a missing tree is the passing case
		}
		if !d.IsDir() {
			found = append(found, path)
		}
		return nil
	})
	return found
}
