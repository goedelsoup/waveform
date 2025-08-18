#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

set -e

echo "Setting up pre-commit for waveform project..."

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "Installing pre-commit..."
    if command -v pip3 &> /dev/null; then
        pip3 install pre-commit
    elif command -v pip &> /dev/null; then
        pip install pre-commit
    else
        echo "Error: pip not found. Please install Python and pip first."
        exit 1
    fi
else
    echo "pre-commit already installed"
fi

# Install pre-commit hooks
echo "Installing pre-commit hooks..."
pre-commit install

# Install golangci-lint if not present
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
else
    echo "golangci-lint already installed"
fi

# Run pre-commit on all files to set up baseline
echo "Running pre-commit on all files..."
pre-commit run --all-files

echo "✅ pre-commit setup complete!"
echo ""
echo "Next steps:"
echo "1. Make sure your Go environment is properly configured"
echo "2. Run 'pre-commit run --all-files' to check all files"
echo "3. The hooks will now run automatically on each commit"
echo ""
echo "Useful commands:"
echo "  pre-commit run --all-files    # Run all hooks on all files"
echo "  pre-commit run                # Run hooks on staged files"
echo "  pre-commit run <hook-id>      # Run a specific hook"
echo "  pre-commit clean              # Clean up pre-commit cache"
