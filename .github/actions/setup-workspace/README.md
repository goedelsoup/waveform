# Setup Workspace Action

This action sets up the development environment for the waveform project, a Go-based OpenTelemetry testing framework.

## Usage

### Basic Usage

```yaml
- name: Setup workspace
  uses: ./.github/actions/setup-workspace
```

### Advanced Usage

```yaml
- name: Setup workspace
  uses: ./.github/actions/setup-workspace
  with:
    go-version: '1.23.3'
    cache-go-modules: 'true'
    install-just: 'true'
    install-tools: 'true'
    setup-git: 'true'
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `go-version` | Go version to use | No | `1.23.3` |
| `cache-go-modules` | Whether to cache Go modules | No | `true` |
| `install-just` | Whether to install just command runner | No | `true` |
| `install-tools` | Whether to install development tools (golangci-lint, godoc) | No | `true` |
| `setup-git` | Whether to configure git for the workflow | No | `true` |

## What This Action Does

1. **Checkout Code**: Checks out the repository with full history
2. **Setup Go**: Installs the specified Go version and optionally caches modules
3. **Install just**: Installs the just command runner for task management
4. **Install Development Tools**: Installs golangci-lint and godoc
5. **Setup Git**: Configures git user for the workflow
6. **Verify Installation**: Verifies Go installation and environment
7. **Download Modules**: Downloads and verifies Go modules
8. **Validate Project**: Runs project structure validation
9. **Show Info**: Displays workspace information and available commands

## Example Workflows

### CI Workflow

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Setup workspace
        uses: ./.github/actions/setup-workspace

      - name: Run tests
        run: just test

      - name: Run linting
        run: just lint
```

### Release Workflow

```yaml
name: Release

on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Setup workspace
        uses: ./.github/actions/setup-workspace
        with:
          install-tools: 'false'  # Skip tools for faster builds

      - name: Build for all platforms
        run: just build-all
```

### Development Workflow

```yaml
name: Development

on: [push]

jobs:
  dev:
    runs-on: ubuntu-latest
    steps:
      - name: Setup workspace
        uses: ./.github/actions/setup-workspace

      - name: Run development workflow
        run: just dev
```

## Available just Commands

After running this action, you can use the following just commands:

- `just build` - Build the binary
- `just test` - Run tests
- `just lint` - Run linting
- `just fmt` - Format code
- `just dev` - Development workflow (fmt, lint, test, build)
- `just ci` - CI workflow (clean, deps, test, build)
- `just release` - Release workflow (clean, deps, test, build-all)
- `just validate` - Validate project structure
- `just check` - Check for common issues

## Notes

- This action uses composite actions, which require GitHub Actions v2
- The action automatically detects if tools are already installed to avoid redundant installations
- Go modules are cached by default to speed up subsequent runs
- The action validates the project structure using the `just validate` command
