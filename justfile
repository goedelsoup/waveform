# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

# Default recipe to run when no arguments are provided
default:
    @just --list

# Build variables
binary_name := "waveform"
build_dir := "build"
version := env_var_or_default("VERSION", "dev")

# Build the binary
build:
    #!/usr/bin/env bash
    echo "Building {{binary_name}}..."
    mkdir -p {{build_dir}}
    go build -ldflags "-X main.Version={{version}}" -o {{build_dir}}/{{binary_name}} ./cmd/waveform

# Install the binary
install:
    #!/usr/bin/env bash
    echo "Installing {{binary_name}}..."
    go install -ldflags "-X main.Version={{version}}" ./cmd/waveform

# Run tests
test:
    #!/usr/bin/env bash
    echo "Running tests..."
    go test -v ./...

# Run tests with coverage
test-coverage:
    #!/usr/bin/env bash
    echo "Running tests with coverage..."
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report generated: coverage.html"

# Run tests for specific package
test-package package:
    #!/usr/bin/env bash
    echo "Running tests for package: {{package}}"
    go test -v ./{{package}}

# Clean build artifacts
clean:
    #!/usr/bin/env bash
    echo "Cleaning build artifacts..."
    rm -rf {{build_dir}}
    rm -f coverage.out coverage.html

# Lint code
lint:
    #!/usr/bin/env bash
    echo "Linting code..."
    golangci-lint run

# Format code
fmt:
    #!/usr/bin/env bash
    echo "Formatting code..."
    go fmt ./...

# Vet code
vet:
    #!/usr/bin/env bash
    echo "Vetting code..."
    go vet ./...

# Generate documentation
docs:
    #!/usr/bin/env bash
    echo "Generating documentation..."
    godoc -http=:6060

# Run example contracts
example: build
    #!/usr/bin/env bash
    echo "Running example contracts..."
    ./{{build_dir}}/{{binary_name}} --contracts "./examples/contracts/**/*.yaml" --verbose

# Run example with reports
example-reports: build
    #!/usr/bin/env bash
    echo "Running example contracts with reports..."
    ./{{build_dir}}/{{binary_name}} \
        --contracts "./examples/contracts/**/*.yaml" \
        --junit-output test-results.xml \
        --lcov-output coverage.info \
        --summary-output summary.txt \
        --verbose

# Test configuration loading
test-config: build
    #!/usr/bin/env bash
    echo "Testing configuration loading..."
    ./{{build_dir}}/{{binary_name}} \
        --contracts "./examples/contracts/**/*.yaml" \
        --config "./examples/collector-config.yaml" \
        --verbose

# Validate configuration file
validate-config:
    #!/usr/bin/env bash
    echo "Validating configuration file..."
    go run ./cmd/waveform --config "./examples/collector-config.yaml" --contracts "./examples/contracts/**/*.yaml" --verbose

# Run end-to-end tests
test-e2e:
    #!/usr/bin/env bash
    echo "Running end-to-end tests..."
    go test -v ./test/e2e/...

# Run all tests including e2e
test-all: test test-e2e

# Build for different platforms
build-linux:
    #!/usr/bin/env bash
    echo "Building for Linux..."
    GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version={{version}}" -o {{build_dir}}/{{binary_name}}-linux-amd64 ./cmd/waveform

build-darwin:
    #!/usr/bin/env bash
    echo "Building for macOS..."
    GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version={{version}}" -o {{build_dir}}/{{binary_name}}-darwin-amd64 ./cmd/waveform

build-windows:
    #!/usr/bin/env bash
    echo "Building for Windows..."
    GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version={{version}}" -o {{build_dir}}/{{binary_name}}-windows-amd64.exe ./cmd/waveform

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Check dependencies
deps:
    #!/usr/bin/env bash
    echo "Checking dependencies..."
    go mod tidy
    go mod verify

# Update dependencies
deps-update:
    #!/usr/bin/env bash
    echo "Updating dependencies..."
    go get -u ./...
    go mod tidy

# Run the CLI tool with default arguments
run: build
    #!/usr/bin/env bash
    echo "Running {{binary_name}} with test contracts..."
    ./{{build_dir}}/{{binary_name}} --contracts "./testdata/test_contracts.yaml" --verbose

# Development workflow: format, lint, test, build
dev: fmt lint test build

# CI workflow: clean, deps, test, build
ci: clean deps test build

# Release workflow: clean, deps, test, build-all
release: clean deps test build-all

# Show help
help:
    @just --list

# Show detailed help
help-detailed:
    @just --list --unsorted

# Validate the project structure
validate:
    #!/usr/bin/env bash
    echo "Validating project structure..."
    
    # Check if required directories exist
    for dir in cmd internal pkg examples testdata; do
        if [ ! -d "$dir" ]; then
            echo "❌ Missing required directory: $dir"
            exit 1
        fi
    done
    
    # Check if required files exist
    for file in go.mod go.sum README.md LICENSE; do
        if [ ! -f "$file" ]; then
            echo "❌ Missing required file: $file"
            exit 1
        fi
    done
    
    echo "✅ Project structure is valid"

# Check for common issues
check:
    #!/usr/bin/env bash
    echo "Running project checks..."
    
    # Check for TODO comments
    echo "Checking for TODO comments..."
    if grep -r "TODO" . --exclude-dir=.git --exclude=justfile; then
        echo "⚠️  Found TODO comments"
    else
        echo "✅ No TODO comments found"
    fi
    
    # Check for FIXME comments
    echo "Checking for FIXME comments..."
    if grep -r "FIXME" . --exclude-dir=.git --exclude=justfile; then
        echo "⚠️  Found FIXME comments"
    else
        echo "✅ No FIXME comments found"
    fi
    
    # Check for debug prints
    echo "Checking for debug prints..."
    if grep -r "fmt.Print" . --exclude-dir=.git --exclude=justfile; then
        echo "⚠️  Found fmt.Print statements"
    else
        echo "✅ No fmt.Print statements found"
    fi

# Install development tools
install-tools:
    #!/usr/bin/env bash
    echo "Installing development tools..."
    
    # Install golangci-lint if not present
    if ! command -v golangci-lint &> /dev/null; then
        echo "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    else
        echo "golangci-lint already installed"
    fi
    
    # Install godoc if not present
    if ! command -v godoc &> /dev/null; then
        echo "Installing godoc..."
        go install golang.org/x/tools/cmd/godoc@latest
    else
        echo "godoc already installed"
    fi
    
    echo "✅ Development tools installed"

# Show project statistics
stats:
    #!/usr/bin/env bash
    echo "Project Statistics:"
    echo "=================="
    
    # Count lines of code
    echo "Lines of Go code:"
    find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | xargs wc -l | tail -1
    
    # Count test files
    echo "Test files:"
    find . -name "*_test.go" -not -path "./vendor/*" -not -path "./.git/*" | wc -l
    
    # Count YAML files
    echo "YAML files:"
    find . -name "*.yaml" -o -name "*.yml" | wc -l
    
    # Count total files
    echo "Total files:"
    find . -type f -not -path "./vendor/*" -not -path "./.git/*" | wc -l
