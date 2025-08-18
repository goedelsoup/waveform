# Evaluate OpenTelemetry Contracts Action

This GitHub Action validates OpenTelemetry contracts against collector pipeline configurations. It's designed to run in pipeline build repositories to ensure that telemetry transformations meet the defined contracts.

## Usage

### Basic Usage

```yaml
- name: Evaluate Contracts
  uses: goedelsoup/waveform/actions/evaluate-contracts@v1
  with:
    contracts: './contracts/**/*.yaml'
    config: './collector-config.yaml'
```

### Advanced Usage

```yaml
- name: Evaluate Contracts
  uses: goedelsoup/waveform/actions/evaluate-contracts@v1
  with:
    contracts: './contracts/**/*.yaml'
    config: './collector-config.yaml'
    mode: 'pipeline'
    junit-output: 'test-results.xml'
    lcov-output: 'coverage.info'
    summary-output: 'summary.txt'
    verbose: 'true'
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `contracts` | Path to contract files (glob patterns supported) | Yes | `./contracts/**/*.yaml` |
| `config` | Path to collector configuration file | No | - |
| `mode` | Test mode: pipeline or processor | No | `pipeline` |
| `junit-output` | Path for JUnit XML output file | No | `test-results.xml` |
| `lcov-output` | Path for LCOV output file | No | `coverage.info` |
| `summary-output` | Path for summary output file | No | `summary.txt` |
| `verbose` | Enable verbose logging | No | `false` |

## Outputs

The action produces several outputs:

- **Test Results**: JUnit XML format for CI/CD integration
- **Coverage Report**: LCOV format for code coverage analysis
- **Summary Report**: Human-readable test summary
- **Artifacts**: All reports are uploaded as GitHub artifacts

## Example Workflow

```yaml
name: Contract Testing

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test-contracts:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Evaluate Contracts
        uses: goedelsoup/waveform/actions/evaluate-contracts@v1
        with:
          contracts: './contracts/**/*.yaml'
          config: './collector-config.yaml'
          verbose: 'true'

      - name: Check Test Results
        if: failure()
        run: |
          echo "Contract tests failed. Check the uploaded artifacts for details."
```

## Integration with Pipeline Builds

This action is typically used in:

1. **Collector Pipeline Repositories**: To validate that pipeline configurations meet service contracts
2. **Infrastructure Repositories**: To ensure telemetry transformations are correctly configured
3. **CI/CD Pipelines**: To catch contract violations before deployment

## Best Practices

1. **Run Early**: Execute contract tests early in your CI pipeline
2. **Use Specific Paths**: Specify exact contract paths rather than broad glob patterns
3. **Enable Verbose Logging**: Use verbose mode for debugging contract issues
4. **Review Artifacts**: Check uploaded artifacts for detailed test results
5. **Fail Fast**: Configure the action to fail the build on contract violations
