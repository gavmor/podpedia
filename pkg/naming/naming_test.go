package naming_test

import (
	"testing"

	"github.com/gavmor/podpedia/pkg/naming"
)

func TestSlug(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"alphanumeric pass-through", "hello", "hello"},
		{"mixed case and numbers", "Hello123", "Hello123"},
		{"hyphen to underscore", "abc-def", "abc_def"},
		{"all caps", "ABC", "ABC"},
		{"non-alphanumeric to underscore", "http://id/123", "http___id_123"},
		{"empty is unknown", "", "unknown"},
		{"multi-byte rune to single underscore", "é", "_"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := naming.Slug(tc.in); got != tc.want {
				t.Errorf("Slug(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNames(t *testing.T) {
	const id = "https://example.com/ep-1"
	const prefix = "https___example_com_ep_1"

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"raw", naming.RawName(id), prefix + "_raw.txt"},
		{"meta", naming.MetaName(id), prefix + "_meta.json"},
		{"audio", naming.AudioName(id), prefix + ".mp3"},
		{"structured", naming.StructuredName(id, "entry"), prefix + "_entry.json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestGlobs(t *testing.T) {
	cases := []struct{ got, want string }{
		{naming.RawGlob, "*_raw.txt"},
		{naming.MetaGlob, "*_meta.json"},
		{naming.AudioGlob, "*.mp3"},
		{naming.StructuredGlob("entry"), "*_entry.json"},
		{naming.StructuredSuffix("entry"), "_entry.json"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}
