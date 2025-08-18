# GitHub Actions

This directory contains reusable GitHub Actions for the waveform project.

## Available Actions

### [setup-workspace](./setup-workspace/)

Sets up the development environment for the waveform project. This action handles:

- Go installation and configuration
- just command runner installation
- Development tools installation (golangci-lint, godoc)
- Git configuration
- Go module downloading and verification
- Project structure validation

**Usage:**
```yaml
- name: Setup workspace
  uses: ./.github/actions/setup-workspace
```

See [setup-workspace/README.md](./setup-workspace/README.md) for detailed documentation.

## Example Workflows

### CI Workflow
See [../workflows/ci.yml](../workflows/ci.yml) for a complete CI workflow example.

### Release Workflow
See [../workflows/release.yml](../workflows/release.yml) for a complete release workflow example.

## Best Practices

1. **Use the setup-workspace action** as the first step in your workflows
2. **Customize inputs** based on your workflow needs (e.g., skip tools installation for faster builds)
3. **Use just commands** for consistent task execution across workflows
4. **Cache artifacts** when appropriate to speed up subsequent runs

## Adding New Actions

When adding new actions:

1. Create a new directory under `.github/actions/`
2. Include an `action.yml` file with proper inputs and outputs
3. Add comprehensive documentation in a `README.md` file
4. Update this README to include the new action
5. Create example workflows demonstrating usage
