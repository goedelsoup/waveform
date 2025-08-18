# Waveform GitHub Actions

This directory contains GitHub Actions for integrating Waveform into CI/CD pipelines. The actions provide a seamless way to validate OpenTelemetry contracts and ensure telemetry pipeline compliance.

## Actions Overview

### 1. Register Contracts Action
**Location**: `actions/register-contracts/`

Used in **application repositories** to:
- Validate contract syntax and structure
- Register contracts for pipeline consumption
- Create contract manifests with metadata
- Upload contracts as GitHub artifacts

### 2. Evaluate Contracts Action
**Location**: `actions/evaluate-contracts/`

Used in **pipeline build repositories** to:
- Download and validate contracts against collector configurations
- Run contract tests against actual pipeline configurations
- Generate test reports (JUnit XML, LCOV, summary)
- Upload test results as artifacts

## Workflow Integration

### Application Repository Workflow

```yaml
name: Register Telemetry Contracts

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  validate-contracts:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Validate Contracts (PR)
        if: github.event_name == 'pull_request'
        uses: goedelsoup/waveform/actions/register-contracts@v1
        with:
          contracts: './contracts/**/*.yaml'
          validate-only: 'true'
          verbose: 'true'

      - name: Register Contracts (Main)
        if: github.event_name == 'push' && github.ref == 'refs/heads/main'
        uses: goedelsoup/waveform/actions/register-contracts@v1
        with:
          contracts: './contracts/**/*.yaml'
          output-dir: './published-contracts'
          verbose: 'true'
```

### Pipeline Repository Workflow

```yaml
name: Validate Telemetry Pipeline

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test-pipeline:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Download Service Contracts
        uses: actions/download-artifact@v3
        with:
          name: registered-contracts
          path: ./service-contracts
          # Note: This requires cross-repository artifact access
          # or manual contract management

      - name: Evaluate Contracts
        uses: goedelsoup/waveform/actions/evaluate-contracts@v1
        with:
          contracts: './service-contracts/**/*.yaml'
          config: './collector-config.yaml'
          mode: 'pipeline'
          verbose: 'true'

      - name: Check Test Results
        if: failure()
        run: |
          echo "Pipeline validation failed. Review the test results above."
```

## Contract Organization

### Recommended Structure

```
contracts/
├── auth-service/
│   ├── http-trace.yaml          # HTTP request/response traces
│   ├── database-trace.yaml      # Database operation traces
│   └── metrics.yaml             # Service metrics
├── user-service/
│   ├── api-trace.yaml           # API endpoint traces
│   ├── business-metrics.yaml    # Business logic metrics
│   └── error-logs.yaml          # Error logging patterns
└── payment-service/
    ├── transaction-trace.yaml   # Payment processing traces
    ├── security-metrics.yaml    # Security-related metrics
    └── audit-logs.yaml          # Audit logging requirements
```

### Contract Naming Conventions

- Use descriptive names that indicate the service and telemetry type
- Include the signal type in the filename (trace, metric, log)
- Use kebab-case for filenames
- Group related contracts in service-specific directories

## Cross-Repository Integration

### Option 1: GitHub Artifacts (Recommended)

1. **Application repositories** register contracts as artifacts
2. **Pipeline repositories** download artifacts from specific workflow runs
3. Use workflow run IDs or commit SHAs to reference specific contract versions

### Option 2: Contract Registry

1. Store contracts in a dedicated contract registry repository
2. Use git submodules or package managers to include contracts
3. Version contracts using semantic versioning

### Option 3: Manual Contract Management

1. Manually copy contracts between repositories
2. Use scripts to synchronize contract versions
3. Maintain contract version compatibility matrices

## Best Practices

### Contract Development

1. **Start Simple**: Begin with basic telemetry contracts
2. **Incremental Complexity**: Add complexity gradually as needs evolve
3. **Version Control**: Use semantic versioning for contract changes
4. **Documentation**: Include clear descriptions and examples
5. **Testing**: Validate contracts locally before committing

### CI/CD Integration

1. **Early Validation**: Run contract validation in pull requests
2. **Fail Fast**: Configure actions to fail builds on contract violations
3. **Artifact Management**: Use appropriate retention periods for artifacts
4. **Cross-Repository Coordination**: Establish clear processes for contract updates
5. **Monitoring**: Track contract compliance metrics over time

### Pipeline Testing

1. **Comprehensive Coverage**: Test all critical telemetry paths
2. **Edge Cases**: Include boundary conditions and error scenarios
3. **Performance**: Consider the impact of contract testing on build times
4. **Maintenance**: Regularly review and update contracts as services evolve

## Troubleshooting

### Common Issues

1. **Contract Loading Errors**: Check YAML syntax and file paths
2. **Validation Failures**: Review contract structure and required fields
3. **Pipeline Mismatches**: Ensure collector configuration matches contract expectations
4. **Artifact Access**: Verify cross-repository artifact permissions

### Debugging

1. Enable verbose logging with `verbose: 'true'`
2. Check action logs for detailed error messages
3. Review uploaded artifacts for test results
4. Validate contracts locally before pushing

## Examples

See the `examples/` directory for complete contract examples and the `test/` directory for integration test scenarios.

## Support

For issues and questions:

1. Check the action documentation in each action directory
2. Review the main Waveform documentation
3. Create an issue on the GitHub repository
4. Check the examples and test files for usage patterns
