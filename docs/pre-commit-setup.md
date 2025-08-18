# Pre-commit Setup

This project uses pre-commit hooks to ensure code quality and consistency before each commit.

## What is pre-commit?

Pre-commit is a framework that manages pre-commit hooks for Git. These hooks run automatically before each commit to check your code for common issues and ensure it meets the project's standards.

## Setup

### Quick Setup

Run the setup script to install and configure pre-commit:

```bash
just setup-pre-commit
```

This will:
1. Install pre-commit if not already installed
2. Install golangci-lint if not already installed
3. Install the pre-commit hooks
4. Run pre-commit on all files to establish a baseline

### Manual Setup

If you prefer to set up manually:

1. Install pre-commit:
   ```bash
   pip install pre-commit
   ```

2. Install golangci-lint:
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

3. Install the hooks:
   ```bash
   pre-commit install
   ```

## Hooks Included

The following hooks are configured:

### Go-specific hooks:
- **go-fmt**: Formats Go code using `gofmt`
- **go-vet**: Runs `go vet` to check for common Go mistakes
- **go-imports**: Organizes and formats imports
- **golangci-lint**: Runs comprehensive linting with golangci-lint
- **go-unit-tests**: Runs unit tests
- **go-build**: Ensures the code compiles
- **go-mod-tidy**: Tidies up go.mod and go.sum

### General file checks:
- **trailing-whitespace**: Removes trailing whitespace
- **end-of-file-fixer**: Ensures files end with a newline
- **check-yaml**: Validates YAML syntax
- **check-added-large-files**: Prevents large files from being committed
- **check-merge-conflict**: Detects merge conflict markers
- **check-case-conflict**: Detects case conflicts in filenames
- **check-json**: Validates JSON syntax
- **debug-statements**: Detects debug statements (fmt.Print, etc.)

### Spell checking:
- **cspell**: Spell checks using the project's cspell configuration

### License headers:
- **check-license-headers**: Ensures Go files have proper license headers

## Usage

### Automatic Usage

Once set up, pre-commit hooks run automatically on every commit. If any hook fails, the commit will be blocked until the issues are fixed.

### Manual Usage

You can also run pre-commit manually:

```bash
# Run on all files
just pre-commit-all

# Run on staged files only
just pre-commit

# Run a specific hook
pre-commit run go-fmt

# Clean the pre-commit cache
just pre-commit-clean
```

## Configuration Files

- `.pre-commit-config.yaml`: Main pre-commit configuration
- `.golangci.yml`: golangci-lint configuration
- `cspell.config.yaml`: Spell checking configuration

## Troubleshooting

### Hook fails on existing code

If hooks fail on existing code that you haven't modified:

1. Run pre-commit on all files to fix issues:
   ```bash
   just pre-commit-all
   ```

2. Commit the fixes, then try your original commit again.

### Performance issues

If pre-commit is slow:

1. Clean the cache:
   ```bash
   just pre-commit-clean
   ```

2. Consider running hooks only on changed files by using `pre-commit run` instead of `pre-commit run --all-files`.

### Skipping hooks (not recommended)

In rare cases, you can skip hooks:

```bash
git commit --no-verify
```

**Warning**: This bypasses all quality checks and should only be used in emergencies.

## Adding New Hooks

To add new hooks, edit `.pre-commit-config.yaml` and run:

```bash
pre-commit install
```

## Best Practices

1. **Fix issues locally**: Always fix pre-commit issues before pushing
2. **Run on all files periodically**: Use `just pre-commit-all` to catch issues in untouched files
3. **Keep hooks updated**: Update pre-commit and hook versions regularly
4. **Customize carefully**: Only modify hooks if absolutely necessary for the project
