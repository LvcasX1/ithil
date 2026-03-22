# Publishing Guide

This guide explains how Ithil is published to package managers via the GitHub Actions release workflow.

## Release Workflow

The release is fully automated via `.github/workflows/rust-release.yml`. When you push a version tag, the workflow:

1. Creates a GitHub release with auto-generated notes
2. Builds release binaries for all platforms (Linux x86_64/ARM64, macOS x86_64/ARM64, Windows x86_64)
3. Uploads tarballs/zips with SHA256 checksums to the release
4. Publishes to crates.io
5. Updates the Homebrew tap formula (`lvcasx1/homebrew-tap`)
6. Updates the AUR package (`ithil-bin`)

Prerelease tags (containing `-`, e.g. `v0.3.0-beta`) skip package manager publishing.

## Creating a Release

```bash
# Create and push a new tag
git tag -a v0.3.0 -m "Release v0.3.0"
git push origin v0.3.0
```

## GitHub Secrets

The following secrets must be configured in `Settings > Secrets and variables > Actions`:

| Secret | Purpose |
|--------|---------|
| `CARGO_REGISTRY_TOKEN` | Publish to crates.io |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Push formula to `lvcasx1/homebrew-tap` (needs `repo` scope) |
| `AUR_SSH_PRIVATE_KEY` | Push PKGBUILD to `aur.archlinux.org/ithil-bin.git` (no passphrase) |

## External Repositories

### Homebrew Tap

Repository: `lvcasx1/homebrew-tap`

The formula at `Formula/ithil.rb` is auto-generated on each release with platform-specific URLs and SHA256 checksums for macOS (x86_64, ARM64) and Linux (x86_64, ARM64).

### AUR Package

Package: `ithil-bin` on [aur.archlinux.org](https://aur.archlinux.org/packages/ithil-bin)

The `PKGBUILD` and `.SRCINFO` are auto-generated and pushed on each release with the Linux x86_64 binary.

## Installation Instructions

### Homebrew (macOS/Linux)
```bash
brew tap lvcasx1/tap
brew install ithil
```

### AUR (Arch Linux)
```bash
yay -S ithil-bin
# or
paru -S ithil-bin
```

### Cargo
```bash
cargo install ithil
```

### Direct Download

Download pre-built binaries from the [releases page](https://github.com/lvcasx1/ithil/releases).

## Troubleshooting

### Homebrew publish fails
- Verify `HOMEBREW_TAP_GITHUB_TOKEN` has `repo` scope
- Ensure `homebrew-tap` repository exists and is accessible

### AUR publish fails
- Verify SSH key has no passphrase
- Test SSH access: `ssh -T aur@aur.archlinux.org`
- Ensure package name `ithil-bin` doesn't conflict with existing packages

### crates.io publish fails
- Verify `CARGO_REGISTRY_TOKEN` is valid
- Ensure version in `Cargo.toml` hasn't been published before
