# Implementation Plan: Update GitHub Actions CI/CD for Tests, Build, and Release

## Phase 1: CI Pipeline Scaffolding [checkpoint: af12f57]
- [x] Task: Create \`.github/workflows/ci.yml\` for automated testing 88cce24
    - [x] Write: Define the workflow to run \`go test\` and \`go lint\`. 88cce24
    - [x] Verify: Run the workflow (simulated or via push) and ensure it fails on broken code and passes on clean code. 88cce24
- [x] Task: Integrate Security Scanning (\`govulncheck\`) fdc7c42
    - [x] Write: Add security scan step to the CI workflow. fdc7c42
    - [x] Verify: Ensure the scan runs correctly. fdc7c42
- [x] Task: Conductor - User Manual Verification 'Phase 1: CI Pipeline Scaffolding' (Protocol in workflow.md)


## Phase 2: GoReleaser & CD Pipeline
- [x] Task: Configure GoReleaser ffec0c4 216ed90
    - [x] Write: Create \`.goreleaser.yaml\` to define the build for Linux (amd64). ffec0c4
    - [x] Verify: Run \`goreleaser build --snapshot --clean\` locally to verify the build process. ffec0c4
- [x] Task: Create \`.github/workflows/release.yml\` for automated distribution fe479c6
    - [x] Write: Define the workflow to run GoReleaser on pushes to \`main\`. fe479c6
    - [x] Verify: Ensure the workflow triggers correctly and handles authentication (GITHUB_TOKEN). fe479c6
- [ ] Task: Conductor - User Manual Verification 'Phase 2: GoReleaser & CD Pipeline' (Protocol in workflow.md)
