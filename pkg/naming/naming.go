// Package naming defines the filename contract for podpedia engine output.
//
// Every artifact the engine writes for an episode shares one filename-safe
// prefix derived from the episode identifier (the RSS <guid>, falling back to
// the audio enclosure URL). A consumer that reads an engine workspace — the
// podpedia-app backend, for instance — has to derive the same prefix to
// correlate output files back to feed items.
//
// That makes the prefix a published contract rather than an engine detail, so
// it lives here instead of behind internal/: a second implementation of Slug
// would be connascence of algorithm across module boundaries, where a change
// on one side still compiles on the other and only diverges at runtime.
package naming

import "strings"

// Unknown is the prefix used for an episode with an empty identifier.
const Unknown = "unknown"

// Suffixes the engine appends to an episode prefix. StructuredSuffix covers
// the extraction output, whose name varies with the scheme ID.
const (
	RawSuffix   = "_raw.txt"
	MetaSuffix  = "_meta.json"
	AudioSuffix = ".mp3"
)

// Glob patterns matching every episode's artifact of a given kind within one
// workspace directory. StructuredGlob covers the scheme-dependent output.
const (
	RawGlob   = "*" + RawSuffix
	MetaGlob  = "*" + MetaSuffix
	AudioGlob = "*" + AudioSuffix
)

// Slug normalizes an episode identifier into a filename-safe prefix by
// replacing every character outside [A-Za-z0-9] with an underscore. An empty
// identifier yields Unknown.
func Slug(s string) string {
	if s == "" {
		return Unknown
	}
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)
}

// StructuredSuffix returns the suffix of the structured extraction output for
// schemeID — "_entry.json" for the "entry" scheme, say.
func StructuredSuffix(schemeID string) string {
	return "_" + schemeID + ".json"
}

// StructuredGlob returns the glob matching every episode's structured output
// for schemeID within one workspace directory.
func StructuredGlob(schemeID string) string {
	return "*" + StructuredSuffix(schemeID)
}

// RawName returns the transcript filename for the episode identified by id.
func RawName(id string) string { return Slug(id) + RawSuffix }

// MetaName returns the metadata sidecar filename for the episode identified
// by id.
func MetaName(id string) string { return Slug(id) + MetaSuffix }

// AudioName returns the downloaded-audio filename for the episode identified
// by id.
func AudioName(id string) string { return Slug(id) + AudioSuffix }

// StructuredName returns the structured-output filename for the episode
// identified by id under the extraction scheme schemeID.
func StructuredName(id, schemeID string) string {
	return Slug(id) + StructuredSuffix(schemeID)
}
