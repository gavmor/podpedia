[Skip to content](https://github.com/gavmor/podpedia/pull/4#start-of-content)

You signed in with another tab or window. [Reload](https://github.com/gavmor/podpedia/pull/4) to refresh your session.You signed out in another tab or window. [Reload](https://github.com/gavmor/podpedia/pull/4) to refresh your session.You switched accounts on another tab or window. [Reload](https://github.com/gavmor/podpedia/pull/4) to refresh your session.Dismiss alert

{{ message }}

[gavmor](https://github.com/gavmor)/ **[podpedia](https://github.com/gavmor/podpedia)** Public

- [Notifications](https://github.com/login?return_to=%2Fgavmor%2Fpodpedia) You must be signed in to change notification settings
- [Fork\\
0](https://github.com/login?return_to=%2Fgavmor%2Fpodpedia)
- [Star\\
0](https://github.com/login?return_to=%2Fgavmor%2Fpodpedia)


## Conversation

[![@gavmor](https://avatars.githubusercontent.com/u/606529?s=80&v=4)](https://github.com/gavmor)


Copy link


Copy Markdown

Owner

### ![@gavmor](https://avatars.githubusercontent.com/u/606529?s=48&v=4)**[gavmor](https://github.com/gavmor)**     commented   [yesterdayApr 28, 2026](https://github.com/gavmor/podpedia/pull/4\#issue-4339772945)•   edited by coderabbitai Bot  Loading          \#\#\# Uh oh!        There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

## Summary

- **Pipeline short-circuits** at the top of \`processEpisode\` — if the entry file already exists on disk, the episode is skipped before any download, whisper, or Ollama call. Like \`npm install\` skipping already-installed packages.
- **Atomic writes** in the store plugin: every file goes through a \`.tmp\` → \`os.Rename\` so a crash mid-write never leaves a partial file that looks complete.
- **Skip-if-exists in the store**: \`HandleStructured\` checks before writing the entry JSON, so the store is also idempotent even if called redundantly.
- **\`\_meta.json\` sidecar**: written alongside every entry, containing \`{episode\_id, audio\_url, title, pub\_date}\`. Schema-independent record — downstream consumers use \`audio\_url\` as their idempotency key without parsing the entry JSON or knowing the extraction scheme.

## Downstream

The ergozetesis adapter (\`adapt-podpedia-output.py\`, committed separately) is now a thin mapping layer: read \`\_meta.json\` sidecars → match institution → merge \`\_entry.json\` → write \`enriched\_from:\` frontmatter. No transcription or extraction logic of its own.

## Test plan

- [ ]  \`go test ./plugins/store/...\` — covers meta sidecar presence and idempotent re-call (new cases)
- [ ]  Run pipeline twice on same feed: second run logs \`already-processed-skipping\` for all episodes
- [ ]  Kill process mid-write, confirm no partial entry files accepted as complete on restart

🤖 Generated with [Claude Code](https://claude.com/claude-code)

## Summary by CodeRabbit

- **Bug Fixes**
  - Enhanced handling of missing or empty episode identifiers for more robust operation
  - Prevents reprocessing of episodes that have already been completed
- **Tests**
  - Extended test coverage to validate metadata sidecar creation and verify required metadata fields
  - Added idempotency tests to ensure episodes are processed only once

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

All reactions

[![@gavmor](https://avatars.githubusercontent.com/u/606529?s=80&u=0580d013d8e432c690572bdf72a8d3446a8b5057&v=4)](https://github.com/gavmor)


Copy link


Copy Markdown

OwnerAuthor

### **[gavmor](https://github.com/gavmor)**     commented   [3 hours agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4\#issuecomment-4338238179)

|     |
| --- |
| [@coderabbitai](https://github.com/coderabbitai) full review |

![eyes](https://github.githubassets.com/assets/1f440-ee44e91e92a7.png)1coderabbitai\[bot\] reacted with eyes emoji

All reactions

- ![eyes](https://github.githubassets.com/assets/1f440-ee44e91e92a7.png)1 reaction

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

[![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=80&v=4)](https://github.com/apps/coderabbitai)


Copy link


Copy Markdown

### **[coderabbitai](https://github.com/apps/coderabbitai) Bot**     commented   [3 hours agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4\#issuecomment-4338239418)

|     |
| --- |
| ✅ Actions performed<br>Full review triggered. |

All reactions

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

[![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=80&v=4)](https://github.com/apps/coderabbitai)


Copy link


Copy Markdown

### **[coderabbitai](https://github.com/apps/coderabbitai) Bot**     commented   [3 hours agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4\#issuecomment-4338241039)•   edited      Loading          \#\#\# Uh oh!        There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

| 📝 Walkthrough

## Walkthrough

The changes implement idempotency mechanisms in the episode processing pipeline. Empty episode IDs are normalized to "unknown", slug-derived entry IDs form consistent file paths, and early-exit checks skip already-processed episodes. Atomic writes replace direct file operations for data integrity, and metadata sidecars are created alongside structured JSON files.

## Changes

| Cohort / File(s) | Summary |
| --- | --- |
| **Pipeline & Storage Idempotency**<br>`internal/pipeline/pipeline.go`, `plugins/store/main.go` | Adds idempotency checks before transcription/audio/extraction; `processEpisode` verifies both structured JSON and metadata sidecar exist before returning early. `HandleStructured` derives paths from slug and scheme\_id, performs stat-based idempotency check, and uses atomic writes for both entry and metadata files. Removes wasm-exported `execute()` function. |
| **Slug Logic**<br>`plugins/store/logic.go` | Normalizes empty input to `"unknown"` instead of producing empty slug. |
| **Test Coverage**<br>`plugins/store/store_test.go` | Validates metadata sidecar creation alongside entry files, checks JSON validity and required keys (`audio_url`, `episode_id`, `title`, `pub_date`). Adds idempotency test confirming second write does not overwrite existing entry file. |
| **Configuration**<br>`.golangci.yml` | Whitespace-only formatting adjustment. |

## Sequence Diagram

Loading

```
sequenceDiagram
    participant Pipeline as internal/pipeline
    participant Storage as plugins/store
    participant FS as Filesystem

    Pipeline->>Pipeline: Normalize empty episode ID to "unknown"
    Pipeline->>Pipeline: Generate slug-derived entryID
    Pipeline->>Storage: Check if already processed
    Storage->>FS: Stat structured JSON + meta sidecar
    FS-->>Storage: Files exist?
    alt Already Processed
        Storage-->>Pipeline: Skip (log: skipping-already-processed)
    else Not Processed
        Pipeline->>Storage: Process episode (transcription, audio, extraction)
        Storage->>Storage: Generate structuredPath from slug(EpisodeID)_scheme_id
        Storage->>FS: Atomic write structured JSON
        Storage->>FS: Atomic write metadata sidecar
        FS-->>Storage: Persistence complete
        Storage-->>Pipeline: Success
    end
```

## Estimated code review effort

🎯 3 (Moderate) \| ⏱️ ~25 minutes

## Poem

> 🐰 Unknown episodes find their name,
>
> Atomic writes bring idempotent fame,
>
> Metadata sidecars skip the twice-told tale,
>
> Where slugs and schemes together sail,
>
> No more rewrites when files prevail! ✨

🚥 Pre-merge checks \| ✅ 4 \| ❌ 1

### ❌ Failed checks (1 warning)

| Check name | Status | Explanation | Resolution |
| :-: | :-- | :-- | :-- |
| Docstring Coverage | ⚠️ Warning | Docstring coverage is 28.57% which is insufficient. The required threshold is 80.00%. | Write docstrings for the functions missing them to satisfy the coverage threshold. |

✅ Passed checks (4 passed)

| Check name | Status | Explanation |
| :-: | :-- | :-- |
| Description Check | ✅ Passed | Check skipped - CodeRabbit’s high-level summary is enabled. |
| Title check | ✅ Passed | The title clearly and concisely captures the two main changes: idempotent episode processing and atomic file writes for durability. |
| Linked Issues check | ✅ Passed | Check skipped because no linked issues were found for this pull request. |
| Out of Scope Changes check | ✅ Passed | Check skipped because no linked issues were found for this pull request. |

✏️ Tip: You can configure your own custom pre-merge checks in the settings.

✨ Finishing Touches📝 Generate docstrings

- [ ]  Create stacked PR
- [ ]  Commit on current branch

🧪 Generate unit tests (beta)

- [ ]  Create PR with unit tests
- [ ]  Commit unit tests in branch `feat/idempotent-runs`

* * *

Comment `@coderabbitai help` to get the list of available commands and usage tips. |

All reactions

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

[![coderabbitai[bot]](https://avatars.githubusercontent.com/in/347564?s=60&v=4)](https://github.com/apps/coderabbitai)

**[coderabbitai](https://github.com/apps/coderabbitai) Bot**

reviewed

[3 hours agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4#pullrequestreview-4191651652)

[View reviewed changes](https://github.com/gavmor/podpedia/pull/4/files)


Copy link


Copy Markdown

### ![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=48&v=4)**[coderabbitai](https://github.com/apps/coderabbitai) Bot**     left a comment

There was a problem hiding this comment.

### Choose a reason for hiding this comment

The reason will be displayed to describe this comment to others. [Learn more](https://docs.github.com/articles/managing-disruptive-comments/#hiding-a-comment).


Choose a reason
SpamAbuseOff TopicOutdatedDuplicateResolvedLow QualityHide comment

**Actionable comments posted: 4**

Caution

Some comments are outside the diff and can’t be posted inline due to platform limitations.

⚠️ Outside diff range comments (1)

> plugins/store/store\_test.go (1)
>
> > `56-78`: _⚠️ Potential issue_ \| _🟡 Minor_
> >
> > **Meta contract assertion is incomplete.**
> >
> > The sidecar contract now includes `episode_id`, `audio_url`, `title`, and `pub_date`, but this test only checks two keys. Please assert all required keys.
> >
> > Proposed test tightening
> >
> > ```
> >  			Expect(meta).To(HaveKey("audio_url"))
> >  			Expect(meta).To(HaveKey("episode_id"))
> > +			Expect(meta).To(HaveKey("title"))
> > +			Expect(meta).To(HaveKey("pub_date"))
> > ```
> >
> > 🤖 Prompt for AI Agents
> >
> > ```
> > Verify each finding against the current code and only fix it if needed.
> >
> > In `@plugins/store/store_test.go` around lines 56 - 78, The meta sidecar assertion
> > in the test inside It("creates the correct entry file and meta sidecar") is
> > incomplete: after reading metaBytes into var meta map[string]string (and
> > unmarshalling), extend the expectations to assert that meta contains the full
> > sidecar contract keys by adding Expect(meta).To(HaveKey("title")) and
> > Expect(meta).To(HaveKey("pub_date")) in addition to the existing checks for
> > "audio_url" and "episode_id"; this change touches the test around
> > HandleStructured, res.Path and metaPath handling.
> > ```

🧹 Nitpick comments (1)

> plugins/store/store\_test.go (1)
>
> > `80-97`: **Idempotency assertion should verify content, not just ModTime.**
> >
> > Using only modification timestamps can miss overwrites on filesystems with coarse timestamp resolution. Prefer asserting file content remains unchanged after the second call.
> >
> > Proposed refactor
> >
> > ```
> >  			entryPath := tmpDir + "/ep_123_my-scheme.json"
> > -			info1, _ := os.Stat(entryPath)
> > +			before, err := os.ReadFile(entryPath)
> > +			Expect(err).NotTo(HaveOccurred())
> >
> >  			// Modify the entry data — second call should NOT overwrite
> >  			req["entry"] = map[string]any{"data": "different data"}
> >  			reqJSON, _ = json.Marshal(req)
> >  			_, err = HandleStructured(reqJSON)
> >  			Expect(err).NotTo(HaveOccurred())
> >
> > -			info2, _ := os.Stat(entryPath)
> > -			Expect(info2.ModTime()).To(Equal(info1.ModTime()))
> > +			after, err := os.ReadFile(entryPath)
> > +			Expect(err).NotTo(HaveOccurred())
> > +			Expect(string(after)).To(Equal(string(before)))
> > ```
> >
> > 🤖 Prompt for AI Agents
> >
> > ```
> > Verify each finding against the current code and only fix it if needed.
> >
> > In `@plugins/store/store_test.go` around lines 80 - 97, The test currently checks
> > idempotency by comparing file ModTime which can be unreliable; change the
> > assertion to read and compare the file contents before and after the second
> > HandleStructured call. In the test block that calls HandleStructured (the "is
> > idempotent" It block), after the first call read the file at entryPath
> > (constructed as tmpDir + "/ep_123_my-scheme.json") into a []byte (or string)
> > variable (e.g., originalBytes), then perform the second HandleStructured call
> > and read the file again into newBytes and assert originalBytes == newBytes
> > (instead of comparing info.ModTime()), ensuring the entry content did not
> > change. Use io.ReadFile or os.ReadFile and retain existing
> > Expect(err).NotTo(HaveOccurred()) checks around reads.
> > ```

🤖 Prompt for all review comments with AI agents

```
Verify each finding against the current code and only fix it if needed.

Inline comments:
In `@internal/pipeline/pipeline.go`:
- Around line 113-120: The idempotency check builds the filename using
slug(ep.ID) but the store logic normalizes empty episode IDs to "unknown",
causing mismatches when ep.ID == "". Update the check to normalize empty IDs the
same way as the store: compute an id variable from ep.ID (e.g., if id == "" set
id = "unknown") and use slug(id) together with the existing suffix logic
(p.schemeID -> suffix) when calling os.Stat so the filename matches the store
path and consistent skipping occurs.
- Around line 117-120: The early return in the block that checks
os.Stat(fmt.Sprintf("%s/%s_%s.json", outputDir, slug(ep.ID), suffix)) causes
processing to stop when the structured file exists but the metadata sidecar is
missing; change the logic in the function around that check (where slug(ep.ID),
suffix and lsess.Info are used) to only skip when both the primary file and its
meta sidecar exist. Concretely, check for the existence of both
fmt.Sprintf("%s/%s_%s.json", outputDir, slug(ep.ID), suffix) and the
corresponding meta file (e.g. fmt.Sprintf("%s/%s_meta.json", outputDir,
slug(ep.ID))) and only call lsess.Info("already-processed-skipping") and return
when both Stats succeed; otherwise fall through to the existing store/refresh
metadata logic so _meta.json can be backfilled.

In `@plugins/store/main.go`:
- Around line 112-116: The current use of a fixed tmp filename (tmp := path +
".tmp") risks collisions on concurrent writes; change this to create a unique
temp file in the same directory (e.g., use os.CreateTemp or os.CreateTemp(dir,
filepath.Base(path)+".tmp-*")), write the data to that temp file (ensuring you
set the desired mode/permissions), call Sync on the file and close it, then
atomically os.Rename the unique temp file to path; update the code paths that
call os.WriteFile/tmp and os.Rename so they use the new temp file handling and
ensure errors are returned if create/write/sync/rename fail.
- Around line 99-101: The code currently swallows failures from
atomicWrite(metaPath, metaJSON) and only logs a warning; change this to
propagate the error so callers know the meta sidecar write failed: when
atomicWrite(metaPath, metaJSON) returns an error, log it with guest.LogMsg
(including err.Error()) and then return or wrap the error from the enclosing
function (use fmt.Errorf or errors.Wrap with context like "failed to write
_meta.json for %s: %w", metaPath, err) instead of returning success; update the
caller handling to surface failure to the client or retry logic so entries are
not left without metadata.

---

Outside diff comments:
In `@plugins/store/store_test.go`:
- Around line 56-78: The meta sidecar assertion in the test inside It("creates
the correct entry file and meta sidecar") is incomplete: after reading metaBytes
into var meta map[string]string (and unmarshalling), extend the expectations to
assert that meta contains the full sidecar contract keys by adding
Expect(meta).To(HaveKey("title")) and Expect(meta).To(HaveKey("pub_date")) in
addition to the existing checks for "audio_url" and "episode_id"; this change
touches the test around HandleStructured, res.Path and metaPath handling.

---

Nitpick comments:
In `@plugins/store/store_test.go`:
- Around line 80-97: The test currently checks idempotency by comparing file
ModTime which can be unreliable; change the assertion to read and compare the
file contents before and after the second HandleStructured call. In the test
block that calls HandleStructured (the "is idempotent" It block), after the
first call read the file at entryPath (constructed as tmpDir +
"/ep_123_my-scheme.json") into a []byte (or string) variable (e.g.,
originalBytes), then perform the second HandleStructured call and read the file
again into newBytes and assert originalBytes == newBytes (instead of comparing
info.ModTime()), ensuring the entry content did not change. Use io.ReadFile or
os.ReadFile and retain existing Expect(err).NotTo(HaveOccurred()) checks around
reads.
```

🪄 Autofix (Beta)

Fix all unresolved CodeRabbit comments on this PR:

- [ ]  Push a commit to this branch (recommended)
- [ ]  Create a new PR with the fixes

* * *

ℹ️ Review info⚙️ Run configuration

**Configuration used**: defaults

**Review profile**: CHILL

**Plan**: Pro Plus

**Run ID**: `c89f5d29-1f3a-40e1-baff-cb4f57037fb7`

📥 Commits

Reviewing files that changed from the base of the PR and between [`a278577`](https://github.com/gavmor/podpedia/commit/a278577f8422bdb0b0bfb129d80cd41699dafaa8) and [`224e104`](https://github.com/gavmor/podpedia/commit/224e104d778e0a368751bcab3bf5541ea3ec9972).

📒 Files selected for processing (3)

- `internal/pipeline/pipeline.go`
- `plugins/store/main.go`
- `plugins/store/store_test.go`

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

All reactions

Comment thread[internal/pipeline/pipeline.go](https://github.com/gavmor/podpedia/pull/4/files#diff-907bfb94a8c2367197d5b3de95c34ecd5cdab2944988165c644bd985722a3c56)
Outdated
Show resolvedHide resolved

### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

Comment thread[internal/pipeline/pipeline.go](https://github.com/gavmor/podpedia/pull/4/files#diff-907bfb94a8c2367197d5b3de95c34ecd5cdab2944988165c644bd985722a3c56)
Outdated
Show resolvedHide resolved

### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

Comment thread[plugins/store/main.go](https://github.com/gavmor/podpedia/pull/4/files#diff-abd6f35b3d4c40f6fbd814efb1d668d1fb94a36c07219ee09d92cfc099f87c11)Show resolvedHide resolved

### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

Comment thread[plugins/store/main.go](https://github.com/gavmor/podpedia/pull/4/files#diff-abd6f35b3d4c40f6fbd814efb1d668d1fb94a36c07219ee09d92cfc099f87c11)
Outdated
Show resolvedHide resolved

### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

[gavmor](https://github.com/gavmor)

added a commit
that referenced
this pull request

[2 hours agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4#ref-commit-caeb7ac)

[![@gavmor](https://avatars.githubusercontent.com/u/606529?s=40&u=0580d013d8e432c690572bdf72a8d3446a8b5057&v=4)](https://github.com/gavmor)

`
          FEAT: Implement idempotent pipeline runs and durable writes\n\nAddres…
`…

`
          caeb7ac
`

```
…ses comments from PR #4:\n- Normalize empty episode IDs to 'unknown' in pipeline and store.\n- Skip processing if both structured data and metadata sidecar exist.\n- Use atomic writes with unique temporary files for concurrency safety.\n- Propagate errors from metadata sidecar writes.
```

[![coderabbitai[bot]](https://avatars.githubusercontent.com/in/347564?s=60&v=4)](https://github.com/apps/coderabbitai)

**[coderabbitai](https://github.com/apps/coderabbitai) Bot**

reviewed

[2 hours agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4#pullrequestreview-4192191647)

[View reviewed changes](https://github.com/gavmor/podpedia/pull/4/files)


Copy link


Copy Markdown

### ![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=48&v=4)**[coderabbitai](https://github.com/apps/coderabbitai) Bot**     left a comment

There was a problem hiding this comment.

### Choose a reason for hiding this comment

The reason will be displayed to describe this comment to others. [Learn more](https://docs.github.com/articles/managing-disruptive-comments/#hiding-a-comment).


Choose a reason
SpamAbuseOff TopicOutdatedDuplicateResolvedLow QualityHide comment

**Actionable comments posted: 1**

🧹 Nitpick comments (3)

> plugins/store/main.go (2)
>
> > `137-148`: **Remove commented-out dead code.**
> >
> > The explanatory comment on line 137 is helpful, but the 12-line commented function body is noise. Consider removing lines 138-148.
> >
> > 🤖 Prompt for AI Agents
> >
> > ```
> > Verify each finding against the current code and only fix it if needed.
> >
> > In `@plugins/store/main.go` around lines 137 - 148, Remove the dead commented-out
> > slug function block from plugins/store/main.go (the 12-line comment body and its
> > surrounding comment markers), leaving only the brief explanatory note that the
> > slug function now lives in logic.go; ensure there are no leftover comment
> > artifacts or extra blank lines so the file stays tidy and builds as before.
> > ```
> >
> > * * *
> >
> > `119-127`: **Consider adding `Sync()` before close for crash durability.**
> >
> > The PR objectives mention preventing partial files on crash. Without `tmpFile.Sync()` before close, data may still be in OS buffers when the rename completes. A power failure at that moment could leave a valid-looking (renamed) file with incomplete content.
> >
> > ♻️ Proposed fix to add fsync
> >
> > ```
> >  	if _, err := tmpFile.Write(content); err != nil {
> >  		tmpFile.Close()
> >  		os.Remove(tmpFile.Name())
> >  		return fmt.Errorf("failed to write to temp file: %w", err)
> >  	}
> > +	if err := tmpFile.Sync(); err != nil {
> > +		tmpFile.Close()
> > +		os.Remove(tmpFile.Name())
> > +		return fmt.Errorf("failed to sync temp file: %w", err)
> > +	}
> >  	if err := tmpFile.Close(); err != nil {
> > ```
> >
> > 🤖 Prompt for AI Agents
> >
> > ```
> > Verify each finding against the current code and only fix it if needed.
> >
> > In `@plugins/store/main.go` around lines 119 - 127, The temp-file write/close
> > sequence should call tmpFile.Sync() after writing and before tmpFile.Close() to
> > ensure data is flushed to disk; update the block that currently uses
> > tmpFile.Write(...) and tmpFile.Close() to call tmpFile.Sync() and check its
> > error (perform the same cleanup: tmpFile.Close() and os.Remove(tmpFile.Name())
> > on error) and return a wrapped fmt.Errorf on Sync failure so crashes won't leave
> > partially-written renamed files.
> > ```
>
> internal/pipeline/pipeline.go (1)
>
> > `256-263`: **Consider aligning `slug()` behavior with `plugins/store/logic.go`.**
> >
> > This `slug()` returns an empty string for empty input, while `plugins/store/logic.go:slug()` returns `"unknown"`. The mismatch is currently safe because `processEpisode` normalizes `ep.ID` beforehand (lines 108-110), but it creates fragility—any future caller that skips normalization will produce mismatched paths.
> >
> > ♻️ Option A: Align behavior
> >
> > ```
> >  func slug(s string) string {
> > +	if s == "" {
> > +		return "unknown"
> > +	}
> >  	return strings.Map(func(r rune) rune {
> > ```
> >
> > ♻️ Option B: Extract shared helper to a common package
> >
> > Move the canonical `slug()` from `plugins/store/logic.go` to a shared internal package and import it in both locations.
> >
> > 🤖 Prompt for AI Agents
> >
> > ```
> > Verify each finding against the current code and only fix it if needed.
> >
> > In `@internal/pipeline/pipeline.go` around lines 256 - 263, The slug function in
> > internal/pipeline/pipeline.go can return an empty string for empty input, which
> > diverges from plugins/store/logic.go's slug that returns "unknown"; update the
> > pipeline.go slug(s string) implementation so that after mapping, if the
> > resulting string is empty return "unknown" (i.e., ensure slug(...) yields
> > "unknown" for empty input), and optionally consider moving the canonical slug
> > into a shared internal package and importing it from both pipeline and
> > plugins/store to avoid future drift.
> > ```

🤖 Prompt for all review comments with AI agents

```
Verify each finding against the current code and only fix it if needed.

Inline comments:
In `@plugins/store/main.go`:
- Around line 39-42: The literal string passed to fmt.Sprintf contains an
unescaped newline which triggers Go's "newline in string" error; update the
fmt.Sprintf call in the assignment to content (used with lo.Coalesce and
req.Episode.Transcript) to use a raw string literal (backticks) or escape the
newline (e.g., "\n\n") when building the "(no transcript)" fallback that
includes req.Episode.Title so the string spans multiple lines legally.

---

Nitpick comments:
In `@internal/pipeline/pipeline.go`:
- Around line 256-263: The slug function in internal/pipeline/pipeline.go can
return an empty string for empty input, which diverges from
plugins/store/logic.go's slug that returns "unknown"; update the pipeline.go
slug(s string) implementation so that after mapping, if the resulting string is
empty return "unknown" (i.e., ensure slug(...) yields "unknown" for empty
input), and optionally consider moving the canonical slug into a shared internal
package and importing it from both pipeline and plugins/store to avoid future
drift.

In `@plugins/store/main.go`:
- Around line 137-148: Remove the dead commented-out slug function block from
plugins/store/main.go (the 12-line comment body and its surrounding comment
markers), leaving only the brief explanatory note that the slug function now
lives in logic.go; ensure there are no leftover comment artifacts or extra blank
lines so the file stays tidy and builds as before.
- Around line 119-127: The temp-file write/close sequence should call
tmpFile.Sync() after writing and before tmpFile.Close() to ensure data is
flushed to disk; update the block that currently uses tmpFile.Write(...) and
tmpFile.Close() to call tmpFile.Sync() and check its error (perform the same
cleanup: tmpFile.Close() and os.Remove(tmpFile.Name()) on error) and return a
wrapped fmt.Errorf on Sync failure so crashes won't leave partially-written
renamed files.
```

🪄 Autofix (Beta)

Fix all unresolved CodeRabbit comments on this PR:

- [ ]  Push a commit to this branch (recommended)
- [ ]  Create a new PR with the fixes

* * *

ℹ️ Review info⚙️ Run configuration

**Configuration used**: defaults

**Review profile**: CHILL

**Plan**: Pro Plus

**Run ID**: `0ab9b62c-b7aa-464b-884d-61c1785ced1b`

📥 Commits

Reviewing files that changed from the base of the PR and between [`224e104`](https://github.com/gavmor/podpedia/commit/224e104d778e0a368751bcab3bf5541ea3ec9972) and [`2278297`](https://github.com/gavmor/podpedia/commit/2278297aa6715bd44d7523cab5f36e72d92f1151).

⛔ Files ignored due to path filters (6)

- `go.sum` is excluded by `!**/*.sum`
- `plugins/downloader/go.sum` is excluded by `!**/*.sum`
- `plugins/extractor/go.sum` is excluded by `!**/*.sum`
- `plugins/rss/go.sum` is excluded by `!**/*.sum`
- `plugins/store/go.sum` is excluded by `!**/*.sum`
- `plugins/transcriber/go.sum` is excluded by `!**/*.sum`

📒 Files selected for processing (3)

- `internal/pipeline/pipeline.go`
- `plugins/store/logic.go`
- `plugins/store/main.go`

✅ Files skipped from review due to trivial changes (1)

- plugins/store/logic.go

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

All reactions

Comment thread[plugins/store/main.go](https://github.com/gavmor/podpedia/pull/4/files#diff-abd6f35b3d4c40f6fbd814efb1d668d1fb94a36c07219ee09d92cfc099f87c11)
Outdated
Show resolvedHide resolved

### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

[![coderabbitai[bot]](https://avatars.githubusercontent.com/in/347564?s=60&v=4)](https://github.com/apps/coderabbitai)

**[coderabbitai](https://github.com/apps/coderabbitai) Bot**

reviewed

[1 hour agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4#pullrequestreview-4192321341)

[View reviewed changes](https://github.com/gavmor/podpedia/pull/4/files)


Copy link


Copy Markdown

### ![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=48&v=4)**[coderabbitai](https://github.com/apps/coderabbitai) Bot**     left a comment

There was a problem hiding this comment.

### Choose a reason for hiding this comment

The reason will be displayed to describe this comment to others. [Learn more](https://docs.github.com/articles/managing-disruptive-comments/#hiding-a-comment).


Choose a reason
SpamAbuseOff TopicOutdatedDuplicateResolvedLow QualityHide comment

**Actionable comments posted: 1**

🤖 Prompt for all review comments with AI agents

```
Verify each finding against the current code and only fix it if needed.

Inline comments:
In `@plugins/store/main.go`:
- Around line 66-71: The current idempotency check in HandleStructured (using
structuredPath := fmt.Sprintf("%s/%s_%s.json", req.OutputDir, entryID,
req.SchemeID)) skips processing as soon as the structured file exists, which
prevents later creation of the associated meta file; change the logic to check
for the meta file as well (e.g., construct the meta path for this entry) and
only skip if both the structured file and its corresponding _meta.json exist—if
structured exists but meta is missing, continue processing so the meta can be
written instead of returning early.
```

🪄 Autofix (Beta)

Fix all unresolved CodeRabbit comments on this PR:

- [ ]  Push a commit to this branch (recommended)
- [ ]  Create a new PR with the fixes

* * *

ℹ️ Review info⚙️ Run configuration

**Configuration used**: defaults

**Review profile**: CHILL

**Plan**: Pro Plus

**Run ID**: `d0c3c7b0-ed67-4f23-a9e1-546a2d45e3d1`

📥 Commits

Reviewing files that changed from the base of the PR and between [`2278297`](https://github.com/gavmor/podpedia/commit/2278297aa6715bd44d7523cab5f36e72d92f1151) and [`e14c8a5`](https://github.com/gavmor/podpedia/commit/e14c8a558bf3614627ad388f299dc4c305fad6a1).

📒 Files selected for processing (4)

- `internal/pipeline/pipeline.go`
- `lefthook.yml`
- `plugins/store/main.go`
- `plugins/store/store_test.go`

🚧 Files skipped from review as they are similar to previous changes (2)

- internal/pipeline/pipeline.go
- plugins/store/store\_test.go

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

All reactions

Comment thread[plugins/store/main.go](https://github.com/gavmor/podpedia/pull/4/files#diff-abd6f35b3d4c40f6fbd814efb1d668d1fb94a36c07219ee09d92cfc099f87c11)Show resolvedHide resolved

### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

[![@gavmor](https://avatars.githubusercontent.com/u/606529?s=40&u=0580d013d8e432c690572bdf72a8d3446a8b5057&v=4)](https://github.com/gavmor)[gavmor](https://github.com/gavmor) [force-pushed](https://github.com/gavmor/podpedia/compare/e14c8a558bf3614627ad388f299dc4c305fad6a1..c6fe168784f2a269c79bf7b2ad5874f585255df4)
the
feat/idempotent-runs
branch
from
[`e14c8a5`](https://github.com/gavmor/podpedia/commit/e14c8a558bf3614627ad388f299dc4c305fad6a1) to
[`c6fe168`](https://github.com/gavmor/podpedia/commit/c6fe168784f2a269c79bf7b2ad5874f585255df4) [Compare](https://github.com/gavmor/podpedia/compare/e14c8a558bf3614627ad388f299dc4c305fad6a1..c6fe168784f2a269c79bf7b2ad5874f585255df4) [14 minutes agoApril 28, 2026 22:26](https://github.com/gavmor/podpedia/pull/4#event-24967874424)

[![@gavmor](https://avatars.githubusercontent.com/u/606529?s=40&v=4)](https://github.com/gavmor)

`
          feat: implement idempotent runs and atomic store operations
` …

Loading

Loading status checks…

### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

`
          429e039
`

```
- Add idempotency checks to skip already processed episodes
- Implement atomic file writes in store plugin to prevent corruption
- Write metadata sidecar (_meta.json) with each entry
- Fix golangci-lint command in lefthook.yml and project scope
- Clean up wasm-microkernel dependencies
- Fix .golangci.yml configuration version mismatch
```

[![@gavmor](https://avatars.githubusercontent.com/u/606529?s=40&u=0580d013d8e432c690572bdf72a8d3446a8b5057&v=4)](https://github.com/gavmor)[gavmor](https://github.com/gavmor) [force-pushed](https://github.com/gavmor/podpedia/compare/c6fe168784f2a269c79bf7b2ad5874f585255df4..429e0391e77eed9e00e57fe4168509034be2e16e)
the
feat/idempotent-runs
branch
from
[`c6fe168`](https://github.com/gavmor/podpedia/commit/c6fe168784f2a269c79bf7b2ad5874f585255df4) to
[`429e039`](https://github.com/gavmor/podpedia/commit/429e0391e77eed9e00e57fe4168509034be2e16e) [Compare](https://github.com/gavmor/podpedia/compare/c6fe168784f2a269c79bf7b2ad5874f585255df4..429e0391e77eed9e00e57fe4168509034be2e16e) [8 minutes agoApril 28, 2026 22:32](https://github.com/gavmor/podpedia/pull/4#event-24968025886)

[![coderabbitai[bot]](https://avatars.githubusercontent.com/in/347564?s=60&v=4)](https://github.com/apps/coderabbitai)

**[coderabbitai](https://github.com/apps/coderabbitai) Bot**

reviewed

[4 minutes agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4#pullrequestreview-4192947568)

[View reviewed changes](https://github.com/gavmor/podpedia/pull/4/files/429e0391e77eed9e00e57fe4168509034be2e16e)


Copy link


Copy Markdown

### ![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=48&v=4)**[coderabbitai](https://github.com/apps/coderabbitai) Bot**     left a comment

There was a problem hiding this comment.

### Choose a reason for hiding this comment

The reason will be displayed to describe this comment to others. [Learn more](https://docs.github.com/articles/managing-disruptive-comments/#hiding-a-comment).


Choose a reason
SpamAbuseOff TopicOutdatedDuplicateResolvedLow QualityHide comment

**Actionable comments posted: 2**

♻️ Duplicate comments (1)

> plugins/store/main.go (1)
>
> > `66-71`: _⚠️ Potential issue_ \| _🟠 Major_
> >
> > **Backfill regression: skip condition still blocks `_meta.json` recovery.**
> >
> > At Line 68, returning early on `structuredPath` existence prevents `_meta.json` creation when only metadata is missing. This contradicts the pipeline’s dual-file skip contract and can leave entries permanently incomplete.
> >
> > Proposed fix
> >
> > ```
> > -	// Idempotency: skip if file exists
> > +	// Idempotency: skip only when both files exist
> >  	structuredPath := fmt.Sprintf("%s/%s_%s.json", req.OutputDir, entryID, req.SchemeID)
> > -	if _, err := os.Stat(structuredPath); err == nil {
> > +	metaPath := fmt.Sprintf("%s/%s_meta.json", req.OutputDir, entryID)
> > +	if _, err := os.Stat(structuredPath); err == nil {
> > +		if _, metaErr := os.Stat(metaPath); metaErr == nil {
> >  		guest.LogMsg("skipping store: " + structuredPath)
> >  		return fmt.Sprintf(`{"path":%q}`, structuredPath), nil
> > +		}
> >  	}
> > @@
> > -	metaPath := fmt.Sprintf("%s/%s_meta.json", req.OutputDir, entryID)
> >  	metaJSON, _ := json.MarshalIndent(meta, "", "  ")
> > ```
> >
> > 🤖 Prompt for AI Agents
> >
> > ```
> > Verify each finding against the current code and only fix it if needed.
> >
> > In `@plugins/store/main.go` around lines 66 - 71, The current idempotency check
> > returns early if structuredPath exists, which blocks recovery when the
> > corresponding meta file (entryID+"_meta.json") is missing; change the logic to
> > compute metaPath (e.g., metaPath := fmt.Sprintf("%s/%s_meta.json",
> > req.OutputDir, entryID)) and only short-circuit (return the JSON path) when both
> > structuredPath and metaPath exist; if structuredPath exists but metaPath does
> > not, proceed to create/write the missing meta file instead of returning early.
> > Ensure you update the condition that uses structuredPath so it checks both
> > files' existence before skipping.
> > ```

🤖 Prompt for all review comments with AI agents

```
Verify each finding against the current code and only fix it if needed.

Inline comments:
In `@plugins/store/main.go`:
- Around line 131-134: The rename block currently moves tmpFile.Name() to path
but does not fsync the parent directory, leaving the directory entry at risk;
after os.Rename(tmpFile.Name(), path) succeeds, open the parent directory (using
filepath.Dir(path)), call Sync on that directory file descriptor, close it, and
return any error wrapped (like fmt.Errorf) so the function fails if the
directory fsync fails; ensure you still remove the temp file on rename error as
before and propagate errors from opening, syncing, or closing the parent
directory.

In `@plugins/store/store_test.go`:
- Around line 73-76: Replace the ignored read error when calling os.ReadFile by
capturing and asserting it so IO failures fail the test explicitly: change the
call that assigns metaBytes to also return err (the symbol metaPath and variable
metaBytes are in scope) and add an assertion (e.g., Expect(err).To(Succeed()) or
Require.NoError) before unmarshalling with json.Unmarshal and asserting on meta;
keep the subsequent json.Unmarshal and Expect(meta).To(HaveKey("audio_url"))
unchanged.

---

Duplicate comments:
In `@plugins/store/main.go`:
- Around line 66-71: The current idempotency check returns early if
structuredPath exists, which blocks recovery when the corresponding meta file
(entryID+"_meta.json") is missing; change the logic to compute metaPath (e.g.,
metaPath := fmt.Sprintf("%s/%s_meta.json", req.OutputDir, entryID)) and only
short-circuit (return the JSON path) when both structuredPath and metaPath
exist; if structuredPath exists but metaPath does not, proceed to create/write
the missing meta file instead of returning early. Ensure you update the
condition that uses structuredPath so it checks both files' existence before
skipping.
```

🪄 Autofix (Beta)

Fix all unresolved CodeRabbit comments on this PR:

- [ ]  Push a commit to this branch (recommended)
- [ ]  Create a new PR with the fixes

* * *

ℹ️ Review info⚙️ Run configuration

**Configuration used**: defaults

**Review profile**: CHILL

**Plan**: Pro Plus

**Run ID**: `126a0f80-f056-4e2e-b00b-96ab12f64024`

📥 Commits

Reviewing files that changed from the base of the PR and between [`e14c8a5`](https://github.com/gavmor/podpedia/commit/e14c8a558bf3614627ad388f299dc4c305fad6a1) and [`429e039`](https://github.com/gavmor/podpedia/commit/429e0391e77eed9e00e57fe4168509034be2e16e).

📒 Files selected for processing (5)

- `.golangci.yml`
- `internal/pipeline/pipeline.go`
- `plugins/store/logic.go`
- `plugins/store/main.go`
- `plugins/store/store_test.go`

✅ Files skipped from review due to trivial changes (2)

- .golangci.yml
- plugins/store/logic.go

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

All reactions

Comment thread[plugins/store/main.go](https://github.com/gavmor/podpedia/pull/4/files/429e0391e77eed9e00e57fe4168509034be2e16e#diff-abd6f35b3d4c40f6fbd814efb1d668d1fb94a36c07219ee09d92cfc099f87c11)

Comment on lines


+131
to
+134


|     |     |     |
| --- | --- | --- |
|  |  | if err := os.Rename(tmpFile.Name(), path); err != nil { |
|  |  | os.Remove(tmpFile.Name()) |
|  |  | return fmt.Errorf("failed to rename temp file: %w", err) |
|  |  | } |


Copy link


Copy Markdown

### ![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=48&v=4)**[coderabbitai](https://github.com/apps/coderabbitai) Bot** [4 minutes agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4\#discussion_r3157626843)

There was a problem hiding this comment.

### Choose a reason for hiding this comment

The reason will be displayed to describe this comment to others. [Learn more](https://docs.github.com/articles/managing-disruptive-comments/#hiding-a-comment).


Choose a reason
SpamAbuseOff TopicOutdatedDuplicateResolvedLow QualityHide comment

_⚠️ Potential issue_ \| _🟠 Major_

🧩 Analysis chain

🏁 Script executed:

```
#!/bin/bash
set -euo pipefail
rg -n -C3 'func atomicWrite|Sync\(|os.Rename\(' plugins/store/main.go
```

Repository: gavmor/podpedia

Length of output: 958

* * *

🏁 Script executed:

```
#!/bin/bash
set -euo pipefail

# Get the complete atomicWrite function
sed -n '104,/^}/p' plugins/store/main.go | head -50

# Check for any comments about durability or crash safety in the file
rg -n 'durable|crash|fsync|sync|atomic' plugins/store/main.go -i
```

Repository: gavmor/podpedia

Length of output: 1607

* * *

🏁 Script executed:

```
#!/bin/bash
set -euo pipefail

# Search for other Sync() calls and file operations
rg -n 'Sync\(\)|WriteFile|os\.Rename' plugins/store/ --type go -B2 -A2

# Check if there are any tests for atomicWrite
fd -e go plugins/store/ | xargs grep -l 'atomicWrite\|Test.*Write' 2>/dev/null || echo "No test files found"
```

Repository: gavmor/podpedia

Length of output: 1020

* * *

**Add parent directory fsync after rename for crash durability.**

The current implementation syncs the temp file's content (line 121) but does not sync the parent directory after the rename (line 131). This creates a crash-safety gap: the file content is durable, but the directory entry may not be persisted on power loss. Since the function documents intent to "prevent data corruption," this should be hardened by syncing the parent directory after the rename.

Proposed fix

```
	if err := os.Rename(tmpFile.Name(), path); err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
+
+	// Ensure directory entry is durable after rename.
+	dirFile, err := os.Open(dir)
+	if err != nil {
+		return fmt.Errorf("failed to open parent directory for sync: %w", err)
+	}
+	defer dirFile.Close()
+	if err := dirFile.Sync(); err != nil {
+		return fmt.Errorf("failed to sync parent directory: %w", err)
+	}

	return nil
}
```

🤖 Prompt for AI Agents

```
Verify each finding against the current code and only fix it if needed.

In `@plugins/store/main.go` around lines 131 - 134, The rename block currently
moves tmpFile.Name() to path but does not fsync the parent directory, leaving
the directory entry at risk; after os.Rename(tmpFile.Name(), path) succeeds,
open the parent directory (using filepath.Dir(path)), call Sync on that
directory file descriptor, close it, and return any error wrapped (like
fmt.Errorf) so the function fails if the directory fsync fails; ensure you still
remove the temp file on rename error as before and propagate errors from
opening, syncing, or closing the parent directory.
```

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

All reactions

Comment thread[plugins/store/store\_test.go](https://github.com/gavmor/podpedia/pull/4/files/429e0391e77eed9e00e57fe4168509034be2e16e#diff-48ff149aaea2926f815b691e0e3844596c4030dad68178a6737b7bc3fc5de6e2)

Comment on lines


+73
to
+76


|     |     |     |
| --- | --- | --- |
|  |  | metaBytes, \_ := os.ReadFile(metaPath) |
|  |  | var meta map\[string\]string |
|  |  | Expect(json.Unmarshal(metaBytes, &meta)).To(Succeed()) |
|  |  | Expect(meta).To(HaveKey("audio\_url")) |


Copy link


Copy Markdown

### ![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=48&v=4)**[coderabbitai](https://github.com/apps/coderabbitai) Bot** [4 minutes agoApr 28, 2026](https://github.com/gavmor/podpedia/pull/4\#discussion_r3157626850)

There was a problem hiding this comment.

### Choose a reason for hiding this comment

The reason will be displayed to describe this comment to others. [Learn more](https://docs.github.com/articles/managing-disruptive-comments/#hiding-a-comment).


Choose a reason
SpamAbuseOff TopicOutdatedDuplicateResolvedLow QualityHide comment

_⚠️ Potential issue_ \| _🟡 Minor_

**Check file-read errors explicitly in the test.**

`metaBytes, _ := os.ReadFile(metaPath)` can mask IO failures and make the JSON assertion misleading.

Proposed fix

```
-			metaBytes, _ := os.ReadFile(metaPath)
+			metaBytes, err := os.ReadFile(metaPath)
+			Expect(err).NotTo(HaveOccurred())
 			var meta map[string]string
 			Expect(json.Unmarshal(metaBytes, &meta)).To(Succeed())
```

📝 Committable suggestion

> ‼️ **IMPORTANT**
>
> Carefully review the code before committing. Ensure that it accurately replaces the highlighted code, contains no missing lines, and has no issues with indentation. Thoroughly test & benchmark the code to ensure it meets the requirements.

Suggested change


|     |     |
| --- | --- |
|  | metaBytes, \_:=os.ReadFile(metaPath) |
|  | varmetamap\[string\]string |
|  | Expect(json.Unmarshal(metaBytes, &meta)).To(Succeed()) |
|  | Expect(meta).To(HaveKey("audio\_url")) |
|  | metaBytes, err:=os.ReadFile(metaPath) |
|  | Expect(err).NotTo(HaveOccurred()) |
|  | varmetamap\[string\]string |
|  | Expect(json.Unmarshal(metaBytes, &meta)).To(Succeed()) |
|  | Expect(meta).To(HaveKey("audio\_url")) |

🤖 Prompt for AI Agents

```
Verify each finding against the current code and only fix it if needed.

In `@plugins/store/store_test.go` around lines 73 - 76, Replace the ignored read
error when calling os.ReadFile by capturing and asserting it so IO failures fail
the test explicitly: change the call that assigns metaBytes to also return err
(the symbol metaPath and variable metaBytes are in scope) and add an assertion
(e.g., Expect(err).To(Succeed()) or Require.NoError) before unmarshalling with
json.Unmarshal and asserting on meta; keep the subsequent json.Unmarshal and
Expect(meta).To(HaveKey("audio_url")) unchanged.
```

Sorry, something went wrong.


### Uh oh!

There was an error while loading. [Please reload this page](https://github.com/gavmor/podpedia/pull/4).

All reactions

This file contains hidden or bidirectional Unicode text that may be interpreted or compiled differently than what appears below. To review, open the file in an editor that reveals hidden Unicode characters.
[Learn more about bidirectional Unicode characters](https://github.co/hiddenchars)

[Show hidden characters](https://github.com/gavmor/podpedia/pull/4)

[Sign up for free](https://github.com/join?source=comment-repo) **to join this conversation on GitHub**.
Already have an account?
[Sign in to comment](https://github.com/login?return_to=https%3A%2F%2Fgithub.com%2Fgavmor%2Fpodpedia%2Fpull%2F4)

### Reviewers

[![@coderabbitai](https://avatars.githubusercontent.com/in/347564?s=40&v=4)](https://github.com/apps/coderabbitai)[coderabbitai\[bot\]](https://github.com/apps/coderabbitai)coderabbitai\[bot\] left review comments

### Assignees

No one assigned

### Labels

None yet

### Projects

None yet

### Milestone

No milestone

### Development

Successfully merging this pull request may close these issues.

None yet

### 1 participant

[![@gavmor](https://avatars.githubusercontent.com/u/606529?s=52&v=4)](https://github.com/gavmor)

Add this suggestion to a batch that can be applied as a single commit.This suggestion is invalid because no changes were made to the code.Suggestions cannot be applied while the pull request is closed.Suggestions cannot be applied while viewing a subset of changes.Only one suggestion per line can be applied in a batch.Add this suggestion to a batch that can be applied as a single commit.Applying suggestions on deleted lines is not supported.You must change the existing code in this line in order to create a valid suggestion.Outdated suggestions cannot be applied.This suggestion has been applied or marked resolved.Suggestions cannot be applied from pending reviews.Suggestions cannot be applied on multi-line comments.Suggestions cannot be applied while the pull request is queued to merge.Suggestion cannot be applied right now. Please check back later.

You can’t perform that action at this time.