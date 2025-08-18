# Development Guide

This document contains development-specific information for contributors to the OpenTelemetry Contract Testing Framework.

## Quick Start

```bash
# Install development tools
just install-tools

# Run development workflow (format, lint, test, build)
just dev

# Run tests
just test

# Build the binary
just build

# Run the CLI tool
just run
```

## Available Commands

```bash
# Show all available commands
just --list

# Build and run examples
just example

# Run with coverage
just test-coverage

# Build for all platforms
just build-all

# Validate project structure
just validate

# Check for common issues
just check

# Show project statistics
just stats
```

## Build System

This project uses [Just](https://github.com/casey/just) as a modern command runner. The `justfile` contains all build, test, and development commands.

Key features:
- **Variable substitution**: Uses `{{variable}}` syntax
- **Dependencies**: Recipes can depend on other recipes
- **Shell scripts**: Supports multi-line shell scripts
- **Cross-platform**: Works on macOS, Linux, and Windows
- **Better error handling**: More informative error messages

## Prerequisites

This project uses [Just](https://github.com/casey/just) as a command runner. Install it with:

```bash
# macOS
brew install just

# Linux
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash

# Windows
scoop install just
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the development workflow: `just dev`
6. Submit a pull request

## License Headers

All source files include SPDX-compliant license headers:

```go
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>
```
