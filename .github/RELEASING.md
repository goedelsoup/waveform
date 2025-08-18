# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

# Releasing Waveform

This document describes how to release new versions of Waveform using GoReleaser.

## Prerequisites

1. **GoReleaser**: Install GoReleaser locally for testing
   ```bash
   go install github.com/goreleaser/goreleaser@latest
   ```

2. **Docker Hub**: Ensure you have Docker Hub credentials configured
   - Create a Docker Hub account if you don't have one
   - Create a repository named `goedelsoup/waveform`

3. **GitHub Secrets**: Set up the following secrets in your GitHub repository:
   - `DOCKERHUB_USERNAME`: Your Docker Hub username
   - `DOCKERHUB_TOKEN`: Your Docker Hub access token

4. **Homebrew Tap**: Create a GitHub repository named `homebrew-tap` in the `goedelsoup` organization

## Release Process

### 1. Prepare for Release

Ensure your code is ready for release:
- All tests pass
- Documentation is up to date
- Version information is correct

### 2. Create a Release Tag

Use conventional commits to automatically generate the next version:

```bash
# For a patch release (bug fixes)
git tag v1.0.1

# For a minor release (new features)
git tag v1.1.0

# For a major release (breaking changes)
git tag v2.0.0
```

### 3. Push the Tag

```bash
git push origin v1.0.1
```

This will automatically trigger the GitHub Actions release workflow.

### 4. Monitor the Release

The release process will:
1. Build binaries for multiple platforms (Linux, macOS, Windows)
2. Create Docker images and push to Docker Hub
3. Create a GitHub release with assets
4. Update the Homebrew formula

## Local Testing

### Test GoReleaser Configuration

```bash
# Test the configuration without publishing
goreleaser release --snapshot --clean --skip-publish

# Check what would be released
goreleaser check
```

### Test Docker Build

```bash
# Build Docker image locally
docker build -t waveform:test .

# Test the image
docker run --rm waveform:test --help
```

## Release Assets

Each release includes:

### Binaries
- `waveform_Linux_x86_64.tar.gz` - Linux AMD64
- `waveform_Linux_arm64.tar.gz` - Linux ARM64
- `waveform_Darwin_x86_64.tar.gz` - macOS AMD64
- `waveform_Darwin_arm64.tar.gz` - macOS ARM64
- `waveform_Windows_x86_64.zip` - Windows AMD64

### Docker Images
- `goedelsoup/waveform:latest` - Latest version
- `goedelsoup/waveform:v1.0.1` - Specific version
- `goedelsoup/waveform:v1.0` - Minor version

### Homebrew
```bash
brew install goedelsoup/tap/waveform
```

## Configuration Files

- `.goreleaser.yml` - GoReleaser configuration
- `Dockerfile` - Docker image definition
- `.github/workflows/release.yml` - GitHub Actions release workflow
- `.github/workflows/test-release.yml` - GitHub Actions test workflow

## Troubleshooting

### Common Issues

1. **Docker Hub Authentication**: Ensure `DOCKERHUB_TOKEN` is set correctly
2. **Homebrew Tap**: Verify the `homebrew-tap` repository exists and is accessible
3. **Git Tags**: Make sure tags follow semantic versioning (v1.0.0, v1.1.0, etc.)

### Manual Release

If the automated release fails, you can run GoReleaser manually:

```bash
# Set required environment variables
export GITHUB_TOKEN=your_github_token
export DOCKER_USERNAME=your_dockerhub_username
export DOCKER_PASSWORD=your_dockerhub_token

# Run the release
goreleaser release --clean
```

## Version Management

Waveform uses semantic versioning (SemVer):
- **Major** (X.0.0): Breaking changes
- **Minor** (0.X.0): New features, backward compatible
- **Patch** (0.0.X): Bug fixes, backward compatible

The version is automatically determined from git tags and conventional commits.
