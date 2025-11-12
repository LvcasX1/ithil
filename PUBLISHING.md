# Publishing Guide

This guide explains how to set up and publish Ithil to multiple package managers using GoReleaser.

## Prerequisites

- GoReleaser installed locally (for testing): `brew install goreleaser` or download from [goreleaser.com](https://goreleaser.com)
- GitHub account with repository access
- Access to package manager platforms (Chocolatey, AUR, etc.)

## GitHub Repositories Setup

### 1. Create Homebrew Tap Repository

```bash
# Create a new repository on GitHub named 'homebrew-tap'
gh repo create lvcasx1/homebrew-tap --public --description "Homebrew tap for Ithil"

# Clone and initialize
git clone https://github.com/lvcasx1/homebrew-tap.git
cd homebrew-tap
mkdir -p Formula
echo "# Homebrew Tap for Ithil" > README.md
git add .
git commit -m "Initial commit"
git push origin main
```

### 2. Create Scoop Bucket Repository

```bash
# Create a new repository on GitHub named 'scoop-bucket'
gh repo create lvcasx1/scoop-bucket --public --description "Scoop bucket for Ithil"

# Clone and initialize
git clone https://github.com/lvcasx1/scoop-bucket.git
cd scoop-bucket
mkdir -p bucket
echo "# Scoop Bucket for Ithil" > README.md
git add .
git commit -m "Initial commit"
git push origin main
```

## GitHub Secrets Configuration

Configure the following secrets in your repository settings (`Settings > Secrets and variables > Actions > New repository secret`):

### 1. HOMEBREW_TAP_GITHUB_TOKEN

**Purpose**: Allows GoReleaser to push Homebrew formula to your tap repository

**Steps**:
1. Go to GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
2. Click "Generate new token (classic)"
3. Name: `GoReleaser Homebrew`
4. Expiration: No expiration (or choose your preference)
5. Scopes: Select `repo` (full control of private repositories)
6. Generate token and copy it
7. Add to repository secrets as `HOMEBREW_TAP_GITHUB_TOKEN`

### 2. SCOOP_GITHUB_TOKEN

**Purpose**: Allows GoReleaser to push Scoop manifest to your bucket repository

**Steps**:
1. Same process as HOMEBREW_TAP_GITHUB_TOKEN
2. Name: `GoReleaser Scoop`
3. Scopes: `repo`
4. Add to repository secrets as `SCOOP_GITHUB_TOKEN`

**Note**: You can reuse the same token for both Homebrew and Scoop, or create separate tokens for better access control.

### 3. CHOCOLATEY_API_KEY

**Purpose**: Allows GoReleaser to publish packages to Chocolatey

**Steps**:
1. Create account at [chocolatey.org](https://chocolatey.org/account/Register)
2. Go to [Account > API Keys](https://community.chocolatey.org/account)
3. Click "Create New API Key"
4. Name: `GoReleaser`
5. Copy the API key
6. Add to repository secrets as `CHOCOLATEY_API_KEY`

**Note**: Initially set `skip_publish: auto` in `.goreleaser.yaml` to test without publishing. Once ready, change to `skip_publish: false`.

### 4. AUR_SSH_PRIVATE_KEY

**Purpose**: Allows GoReleaser to push package to Arch User Repository (AUR)

**Steps**:

1. **Create AUR account** at [aur.archlinux.org](https://aur.archlinux.org/register/)

2. **Generate SSH key for AUR**:
   ```bash
   ssh-keygen -t ed25519 -C "your_email@example.com" -f ~/.ssh/aur
   # Press Enter for no passphrase (required for automation)
   ```

3. **Add public key to AUR account**:
   ```bash
   cat ~/.ssh/aur.pub
   # Copy the output
   ```
   - Go to AUR Account Settings
   - Paste public key in "SSH Public Key" section
   - Save

4. **Test AUR access**:
   ```bash
   ssh -T aur@aur.archlinux.org -i ~/.ssh/aur
   # Should see: "Hi username! You've successfully authenticated..."
   ```

5. **Add private key to GitHub secrets**:
   ```bash
   cat ~/.ssh/aur
   # Copy the ENTIRE output including BEGIN/END lines
   ```
   - Add to repository secrets as `AUR_SSH_PRIVATE_KEY`

6. **Create AUR package placeholder** (required before first publish):
   ```bash
   # Clone the new package repository
   git clone ssh://aur@aur.archlinux.org/ithil-bin.git
   cd ithil-bin

   # Create initial PKGBUILD
   cat > PKGBUILD << 'EOF'
   # Maintainer: LvcasX1 <contact@lvcasx1.dev>
   pkgname=ithil-bin
   pkgver=0.1.0
   pkgrel=1
   pkgdesc="A Terminal User Interface (TUI) Telegram client built with Go and Bubbletea"
   arch=('x86_64')
   url="https://github.com/lvcasx1/ithil"
   license=('MIT')
   provides=('ithil')
   conflicts=('ithil')
   source=("https://github.com/lvcasx1/ithil/releases/download/v${pkgver}/ithil_${pkgver}_Linux_x86_64.tar.gz")
   sha256sums=('SKIP')

   package() {
     install -Dm755 "${srcdir}/ithil" "${pkgdir}/usr/bin/ithil"
   }
   EOF

   # Create .SRCINFO
   makepkg --printsrcinfo > .SRCINFO

   # Commit and push
   git add PKGBUILD .SRCINFO
   git commit -m "Initial package for ithil-bin"
   git push origin master
   ```

## Testing Locally

Before creating a release, test GoReleaser configuration locally:

```bash
# Install GoReleaser
brew install goreleaser

# Test configuration (dry-run, no publishing)
goreleaser check

# Build snapshot (creates binaries without publishing)
goreleaser release --snapshot --clean

# Check the dist/ folder for generated artifacts
ls -la dist/
```

This will:
- Validate `.goreleaser.yaml` configuration
- Build binaries for all platforms
- Generate archives and packages
- Create checksums
- NOT publish anything (snapshot mode)

## Creating a Release

Once everything is configured:

```bash
# Create and push a new tag
git tag -a v0.1.1 -m "Release v0.1.1"
git push origin v0.1.1
```

GitHub Actions will automatically:
1. Trigger the release workflow
2. Run GoReleaser
3. Build binaries for all platforms
4. Create GitHub release with artifacts
5. Publish to Homebrew (if token configured)
6. Publish to Scoop (if token configured)
7. Publish to Chocolatey (if API key configured and skip_publish=false)
8. Publish to AUR (if SSH key configured)
9. Generate DEB, RPM, and APK packages

## Installation Instructions (After Publishing)

### Homebrew (macOS/Linux)
```bash
brew tap lvcasx1/tap
brew install ithil
```

### Scoop (Windows)
```bash
scoop bucket add lvcasx1 https://github.com/lvcasx1/scoop-bucket
scoop install ithil
```

### Chocolatey (Windows)
```bash
choco install ithil
```

### AUR (Arch Linux)
```bash
yay -S ithil-bin
# or
paru -S ithil-bin
```

### DEB (Debian/Ubuntu)
```bash
# Download .deb from GitHub releases
wget https://github.com/lvcasx1/ithil/releases/download/v0.1.0/ithil_0.1.0_Linux_x86_64.deb
sudo dpkg -i ithil_0.1.0_Linux_x86_64.deb
```

### RPM (Fedora/RHEL/CentOS)
```bash
# Download .rpm from GitHub releases
wget https://github.com/lvcasx1/ithil/releases/download/v0.1.0/ithil_0.1.0_Linux_x86_64.rpm
sudo rpm -i ithil_0.1.0_Linux_x86_64.rpm
```

### APK (Alpine Linux)
```bash
# Download .apk from GitHub releases
wget https://github.com/lvcasx1/ithil/releases/download/v0.1.0/ithil_0.1.0_Linux_x86_64.apk
sudo apk add --allow-untrusted ithil_0.1.0_Linux_x86_64.apk
```

## Troubleshooting

### GoReleaser fails with "git is in dirty state"
```bash
git status
# Commit any uncommitted changes
git add .
git commit -m "Update before release"
```

### Homebrew publish fails
- Verify `HOMEBREW_TAP_GITHUB_TOKEN` has `repo` scope
- Ensure `homebrew-tap` repository exists and is accessible
- Check GitHub Actions logs for detailed error

### Chocolatey publish fails
- Verify API key is correct
- First-time packages may require manual approval on Chocolatey
- Consider using `skip_publish: auto` initially

### AUR publish fails
- Verify SSH key has no passphrase
- Test SSH access: `ssh -T aur@aur.archlinux.org`
- Ensure package name `ithil-bin` doesn't conflict with existing packages
- Check that initial PKGBUILD was pushed to AUR

### Build fails for specific platform
- Check Go version compatibility
- Verify CGO_ENABLED=0 for static builds
- Test cross-compilation locally:
  ```bash
  GOOS=windows GOARCH=amd64 go build ./cmd/ithil
  ```

## Maintenance

### Updating Package Descriptions
Edit `.goreleaser.yaml` and modify the description fields for each package manager.

### Adding New Platforms
Add to the `builds` section in `.goreleaser.yaml`:
```yaml
builds:
  - goos:
      - linux
      - darwin
      - windows
      - freebsd  # Add new OS
    goarch:
      - amd64
      - arm64
      - arm  # Add new architecture
```

### Removing a Package Manager
Comment out or remove the relevant section in `.goreleaser.yaml` and remove the corresponding secret from GitHub.

## Security Notes

- **Never commit secrets** to the repository
- Use GitHub's secret management for all API keys and tokens
- Rotate tokens periodically
- Use fine-grained tokens with minimal permissions when possible
- For AUR, ensure SSH private key has no passphrase (required for automation)
- Keep `.goreleaser.yaml` in version control, but ensure it uses environment variables for sensitive data

## Resources

- [GoReleaser Documentation](https://goreleaser.com/intro/)
- [Homebrew Tap Creation](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Scoop Bucket Creation](https://github.com/ScoopInstaller/Scoop/wiki/Buckets)
- [Chocolatey Package Creation](https://docs.chocolatey.org/en-us/create/create-packages)
- [AUR Package Guidelines](https://wiki.archlinux.org/title/AUR_submission_guidelines)
