# Implementation Plan: Update GitHub Actions CI/CD for Tests, Build, and Release

## Phase 1: CI Pipeline Scaffolding
- [x] Task: Create \`.github/workflows/ci.yml\` for automated testing 88cce24
    - [x] Write: Define the workflow to run \`go test\` and \`go lint\`. 88cce24
    - [x] Verify: Run the workflow (simulated or via push) and ensure it fails on broken code and passes on clean code. 88cce24

- [ ] Task: Integrate Security Scanning (`govulncheck`)
    - [ ] Write: Add security scan step to the CI workflow.
    - [ ] Verify: Ensure the scan runs correctly.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: CI Pipeline Scaffolding' (Protocol in workflow.md)

## Phase 2: GoReleaser & CD Pipeline
- [ ] Task: Configure GoReleaser
    - [ ] Write: Create `.goreleaser.yaml` to define the build for Linux (amd64).
    - [ ] Verify: Run `goreleaser build --snapshot --clean` locally to verify the build process.
- [ ] Task: Create `.github/workflows/release.yml` for automated distribution
    - [ ] Write: Define the workflow to run GoReleaser on pushes to `main`.
    - [ ] Verify: Ensure the workflow triggers correctly and handles authentication (GITHUB_TOKEN).
- [ ] Task: Conductor - User Manual Verification 'Phase 2: GoReleaser & CD Pipeline' (Protocol in workflow.md)
