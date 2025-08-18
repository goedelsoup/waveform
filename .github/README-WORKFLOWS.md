# GitHub Workflows

This directory contains GitHub Actions workflows and reusable actions for the OpenTelemetry Contract Testing Framework project.

## Overview

The project uses GitHub Actions for continuous integration, testing, and automated releases. All workflows are designed to be reliable, fast, and provide comprehensive feedback.

## Available Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` branch

**Jobs:**

#### Test Job
- **Purpose**: Run the complete test suite
- **Steps**:
  - Setup workspace environment
  - Run unit tests (`just test`)
  - Run tests with coverage (`just test-coverage`)
  - Upload coverage to Codecov
- **Artifacts**: Coverage reports

#### Lint Job
- **Purpose**: Ensure code quality and consistency
- **Steps**:
  - Setup workspace with development tools
  - Run linting (`just lint`)
  - Check code formatting (`just fmt`)
  - Run Go vet (`just vet`)
- **Artifacts**: None (fails on issues)

#### Build Job
- **Purpose**: Build the binary for distribution
- **Steps**:
  - Setup workspace (without development tools for speed)
  - Build binary (`just build`)
  - Upload build artifacts
- **Artifacts**: Binary files (retained for 7 days)

#### Validate Job
- **Purpose**: Validate project structure and run examples
- **Steps**:
  - Setup workspace
  - Validate project structure (`just validate`)
  - Check for common issues (`just check`)
  - Run example contracts (`just example`)
- **Artifacts**: None (fails on issues)

### 2. Release Workflow (`.github/workflows/release.yml`)

**Triggers:**
- Push of tags matching pattern `v*` (e.g., `v1.0.0`)

**Jobs:**

#### Build Release Job
- **Purpose**: Build binaries for all supported platforms
- **Steps**:
  - Setup workspace (optimized for builds)
  - Build for all platforms (`just build-all`)
  - Upload release artifacts
- **Artifacts**: Multi-platform binaries (retained for 30 days)

#### Test Release Job
- **Purpose**: Validate the release with full test suite
- **Steps**:
  - Setup workspace
  - Run full test suite (`just test`)
  - Run example contracts with reports (`just example-reports`)
  - Upload test results
- **Artifacts**: Test results, coverage, and summary files

#### Create Release Job
- **Purpose**: Create GitHub release with artifacts
- **Dependencies**: Requires both build and test jobs to succeed
- **Steps**:
  - Setup workspace with Git configuration
  - Download build artifacts
  - Download test results
  - Create GitHub release with all artifacts
- **Artifacts**: GitHub release with binaries and test results

## Reusable Actions

### Setup Workspace Action (`.github/actions/setup-workspace/`)

A composite action that sets up the complete development environment.

**Inputs:**
- `go-version`: Go version to use (default: `1.23.3`)
- `cache-go-modules`: Whether to cache Go modules (default: `true`)
- `install-just`: Whether to install just command runner (default: `true`)
- `install-tools`: Whether to install development tools (default: `true`)
- `setup-git`: Whether to configure Git (default: `true`)

**Usage:**
```yaml
- name: Setup workspace
  uses: ./.github/actions/setup-workspace
  with:
    go-version: '1.23.3'
    install-tools: 'false'  # Skip for faster builds
```

**What it does:**
- Checks out code with full history
- Sets up Go environment with caching
- Installs just command runner
- Installs development tools (golangci-lint, godoc)
- Configures Git for workflows
- Downloads and verifies Go modules
- Validates project structure
- Shows workspace information

## Workflow Best Practices

### 1. Use the Setup Workspace Action
Always start your workflows with the setup-workspace action for consistent environment setup:

```yaml
- name: Setup workspace
  uses: ./.github/actions/setup-workspace
```

### 2. Optimize for Speed
For build jobs, skip development tools installation:

```yaml
- name: Setup workspace
  uses: ./.github/actions/setup-workspace
  with:
    install-tools: 'false'
```

### 3. Use Just Commands
All workflows use just commands for consistent task execution:

```yaml
- name: Run tests
  run: just test

- name: Build binary
  run: just build
```

### 4. Proper Artifact Management
- Use appropriate retention periods
- Upload artifacts only when needed
- Download artifacts in dependent jobs

### 5. Comprehensive Testing
- Run tests in multiple jobs for different purposes
- Include coverage reporting
- Validate examples and project structure

## Workflow Triggers

### CI Workflow
```yaml
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]
```

### Release Workflow
```yaml
on:
  push:
    tags: ['v*']
```

## Artifact Retention

- **CI artifacts**: 7 days (build binaries)
- **Release artifacts**: 30 days (all release files)
- **Coverage reports**: Uploaded to Codecov (permanent)

## Monitoring and Debugging

### Workflow Status
- All workflows run on Ubuntu latest
- Jobs run in parallel when possible
- Release workflow requires all jobs to succeed

### Common Issues
1. **Test failures**: Check test output and coverage reports
2. **Build failures**: Verify Go module dependencies
3. **Lint failures**: Run `just lint` locally to fix issues
4. **Validation failures**: Check project structure with `just validate`

### Debugging Tips
- Use verbose logging with `--verbose` flags
- Check artifact contents for detailed information
- Review Codecov reports for test coverage issues
- Use local development workflow: `just dev`

## Adding New Workflows

When adding new workflows:

1. **Follow naming conventions**: Use descriptive names
2. **Use the setup-workspace action**: For consistent environment setup
3. **Include proper triggers**: Define when the workflow should run
4. **Add appropriate jobs**: Separate concerns into different jobs
5. **Handle artifacts properly**: Upload and download as needed
6. **Include documentation**: Update this README

### Example New Workflow
```yaml
name: Security Scan

on:
  schedule:
    - cron: '0 2 * * 1'  # Weekly on Monday at 2 AM
  workflow_dispatch:  # Manual trigger

jobs:
  security:
    name: Security Scan
    runs-on: ubuntu-latest
    
    steps:
      - name: Setup workspace
        uses: ./.github/actions/setup-workspace
      
      - name: Run security scan
        run: just security-scan
      
      - name: Upload results
        uses: actions/upload-artifact@v4
        with:
          name: security-results
          path: security-report.json
```

## Integration with Development

### Local Development
Use the same commands locally that the workflows use:

```bash
# Run the full CI workflow locally
just dev

# Run specific workflow steps
just test
just lint
just build
just validate
```

### Pre-commit Checks
Before pushing, run the CI checks locally:

```bash
# Run all checks
just check

# Run specific checks
just fmt
just lint
just test
```

## Support

For workflow issues:

1. Check the workflow logs for detailed error messages
2. Verify that local development commands work
3. Review the justfile for command definitions
4. Check the setup-workspace action documentation
5. Create an issue with workflow logs and reproduction steps
