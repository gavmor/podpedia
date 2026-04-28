# Specification: Continuous Binary Releases

## Goal
To ensure that every successful build on the `main` branch results in a published binary release on GitHub.

## Requirements
1.  **Trigger:** Successful completion of `test`, `lint`, and `vulncheck` jobs in the `CI` workflow on the `main` branch.
2.  **Versioning:** Automated semantic versioning based on commit messages, or a "rolling" release if preferred. Given the current workflow, automated tagging is recommended.
3.  **Artifacts:** Binaries for Linux and Darwin (AMD64 and ARM64), as currently configured in `.goreleaser.yaml`.
4.  **Process:**
    -   Run standard CI checks.
    -   If checks pass and branch is `main`, determine the next version.
    -   Create and push a new git tag.
    -   Trigger `goreleaser` to build and upload artifacts.

## Success Criteria
-   Pushing a change to `main` results in a new GitHub Release with binaries once CI passes.
-   Releases contain the expected artifacts (tarballs/zips).
-   Version numbers increment correctly.
