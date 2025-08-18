# Register OpenTelemetry Contracts Action

This GitHub Action registers OpenTelemetry contracts from application repositories, making them available for pipeline validation. It's designed to run in application repositories to publish contracts that can be consumed by pipeline build processes.

## Usage

### Basic Usage

```yaml
- name: Register Contracts
  uses: goedelsoup/waveform/actions/register-contracts@v1
  with:
    contracts: './contracts/**/*.yaml'
```

### Advanced Usage

```yaml
- name: Register Contracts
  uses: goedelsoup/waveform/actions/register-contracts@v1
  with:
    contracts: './contracts/**/*.yaml'
    output-dir: './published-contracts'
    validate-only: 'false'
    verbose: 'true'
```

### Validation Only

```yaml
- name: Validate Contracts
  uses: goedelsoup/waveform/actions/register-contracts@v1
  with:
    contracts: './contracts/**/*.yaml'
    validate-only: 'true'
    verbose: 'true'
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `contracts` | Path to contract files (glob patterns supported) | Yes | `./contracts/**/*.yaml` |
| `output-dir` | Directory to store registered contracts | No | `./registered-contracts` |
| `validate-only` | Only validate contracts without registering | No | `false` |
| `verbose` | Enable verbose logging | No | `false` |

## Outputs

The action produces several outputs:

- **Validated Contracts**: All contracts are validated for syntax and structure
- **Registered Contracts**: Contracts are copied to the output directory
- **Contract Manifest**: JSON manifest with metadata about registered contracts
- **Artifacts**: Registered contracts are uploaded as GitHub artifacts

## Example Workflow

```yaml
name: Register Contracts

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  register-contracts:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Register Contracts
        uses: goedelsoup/waveform/actions/register-contracts@v1
        with:
          contracts: './contracts/**/*.yaml'
          output-dir: './published-contracts'
          verbose: 'true'

      - name: Validate Contract Structure
        if: failure()
        run: |
          echo "Contract registration failed. Check the validation output above."
```

## Contract Manifest

The action creates a `manifest.json` file with metadata about the registered contracts:

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "repository": "myorg/myapp",
  "commit": "abc123def456",
  "contracts": [
    "auth-service.yaml",
    "user-service.yaml",
    "payment-service.yaml"
  ]
}
```

## Integration with Application Repositories

This action is typically used in:

1. **Service Repositories**: To register telemetry contracts for each service
2. **Application Repositories**: To publish contracts that define telemetry expectations
3. **Microservice Repositories**: To ensure each service defines its telemetry contracts

## Best Practices

1. **Validate Early**: Use `validate-only: true` in pull requests to catch issues early
2. **Organize Contracts**: Use clear directory structures for contract organization
3. **Version Contracts**: Include version information in contract files
4. **Document Expectations**: Use descriptive contract names and descriptions
5. **Review Artifacts**: Check uploaded artifacts to ensure contracts are properly registered

## Contract Organization

Recommended directory structure:

```
contracts/
├── auth-service/
│   ├── http-trace.yaml
│   └── metrics.yaml
├── user-service/
│   ├── database-trace.yaml
│   └── api-metrics.yaml
└── payment-service/
    ├── transaction-trace.yaml
    └── business-metrics.yaml
```

## Pipeline Integration

Registered contracts can be consumed by the `evaluate-contracts` action in pipeline repositories:

```yaml
# In pipeline repository workflow
- name: Download Contracts
  uses: actions/download-artifact@v3
  with:
    name: registered-contracts
    path: ./downloaded-contracts

- name: Evaluate Contracts
  uses: goedelsoup/waveform/actions/evaluate-contracts@v1
  with:
    contracts: './downloaded-contracts/**/*.yaml'
    config: './collector-config.yaml'
```
