# Cross-Repository Contract Testing Example

This example demonstrates how to use both Waveform GitHub Actions together in a microservices architecture with separate application and pipeline repositories.

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Auth Service  │    │  User Service   │    │ Payment Service │
│   Repository    │    │  Repository     │    │  Repository     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │ Pipeline Build  │
                    │  Repository     │
                    └─────────────────┘
```

## Application Repository Workflows

### Auth Service Repository

**File**: `.github/workflows/register-contracts.yml`

```yaml
name: Register Telemetry Contracts

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

**Contract Structure**:
```
contracts/
├── http-trace.yaml
├── database-trace.yaml
└── metrics.yaml
```

### User Service Repository

**File**: `.github/workflows/register-contracts.yml`

```yaml
name: Register Telemetry Contracts

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

**Contract Structure**:
```
contracts/
├── api-trace.yaml
├── business-metrics.yaml
└── error-logs.yaml
```

## Pipeline Repository Workflow

### Pipeline Build Repository

**File**: `.github/workflows/validate-pipeline.yml`

```yaml
name: Validate Telemetry Pipeline

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:
    inputs:
      auth_service_commit:
        description: 'Auth service commit SHA'
        required: false
      user_service_commit:
        description: 'User service commit SHA'
        required: false
      payment_service_commit:
        description: 'Payment service commit SHA'
        required: false

jobs:
  download-contracts:
    runs-on: ubuntu-latest
    outputs:
      auth-contracts: ${{ steps.download-auth.outputs.path }}
      user-contracts: ${{ steps.download-user.outputs.path }}
      payment-contracts: ${{ steps.download-payment.outputs.path }}
    steps:
      - name: Download Auth Service Contracts
        id: download-auth
        uses: actions/download-artifact@v3
        with:
          name: registered-contracts
          path: ./auth-contracts
          # Note: This requires cross-repository artifact access
          # or manual contract management

      - name: Download User Service Contracts
        id: download-user
        uses: actions/download-artifact@v3
        with:
          name: registered-contracts
          path: ./user-contracts

      - name: Download Payment Service Contracts
        id: download-payment
        uses: actions/download-artifact@v3
        with:
          name: registered-contracts
          path: ./payment-contracts

  validate-pipeline:
    needs: download-contracts
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Evaluate Auth Service Contracts
        uses: goedelsoup/waveform/actions/evaluate-contracts@v1
        with:
          contracts: './auth-contracts/**/*.yaml'
          config: './collector-config.yaml'
          mode: 'pipeline'
          junit-output: 'auth-test-results.xml'
          lcov-output: 'auth-coverage.info'
          summary-output: 'auth-summary.txt'
          verbose: 'true'

      - name: Evaluate User Service Contracts
        uses: goedelsoup/waveform/actions/evaluate-contracts@v1
        with:
          contracts: './user-contracts/**/*.yaml'
          config: './collector-config.yaml'
          mode: 'pipeline'
          junit-output: 'user-test-results.xml'
          lcov-output: 'user-coverage.info'
          summary-output: 'user-summary.txt'
          verbose: 'true'

      - name: Evaluate Payment Service Contracts
        uses: goedelsoup/waveform/actions/evaluate-contracts@v1
        with:
          contracts: './payment-contracts/**/*.yaml'
          config: './collector-config.yaml'
          mode: 'pipeline'
          junit-output: 'payment-test-results.xml'
          lcov-output: 'payment-coverage.info'
          summary-output: 'payment-summary.txt'
          verbose: 'true'

      - name: Generate Combined Report
        run: |
          echo "Combined Contract Test Results" > combined-summary.txt
          echo "=============================" >> combined-summary.txt
          echo "" >> combined-summary.txt
          
          echo "Auth Service:" >> combined-summary.txt
          cat auth-summary.txt >> combined-summary.txt
          echo "" >> combined-summary.txt
          
          echo "User Service:" >> combined-summary.txt
          cat user-summary.txt >> combined-summary.txt
          echo "" >> combined-summary.txt
          
          echo "Payment Service:" >> combined-summary.txt
          cat payment-summary.txt >> combined-summary.txt

      - name: Upload Combined Results
        uses: actions/upload-artifact@v3
        with:
          name: 'combined-contract-results'
          path: |
            combined-summary.txt
            *-test-results.xml
            *-coverage.info
            *-summary.txt
          retention-days: 30

      - name: Check Test Results
        if: failure()
        run: |
          echo "Pipeline validation failed. Check the uploaded artifacts for details."
          echo "Review the combined results in the 'combined-contract-results' artifact."
```

## Alternative: Contract Registry Approach

If cross-repository artifact access is not available, you can use a dedicated contract registry:

### Contract Registry Repository

**File**: `.github/workflows/update-registry.yml`

```yaml
name: Update Contract Registry

on:
  workflow_run:
    workflows: ["Register Telemetry Contracts"]
    types: [completed]

jobs:
  update-registry:
    runs-on: ubuntu-latest
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    steps:
      - name: Checkout registry
        uses: actions/checkout@v4
        with:
          repository: myorg/telemetry-contracts
          token: ${{ secrets.REGISTRY_TOKEN }}

      - name: Download contracts
        uses: actions/download-artifact@v3
        with:
          name: registered-contracts
          path: ./temp-contracts

      - name: Update registry
        run: |
          # Copy contracts to appropriate service directory
          cp -r ./temp-contracts/* ./contracts/${{ github.event.workflow_run.repository.name }}/
          
          # Update manifest
          echo "Updated contracts from ${{ github.event.workflow_run.repository.name }}"

      - name: Commit and push
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add .
          git commit -m "Update contracts from ${{ github.event.workflow_run.repository.name }}"
          git push
```

### Pipeline Repository (Registry Approach)

**File**: `.github/workflows/validate-pipeline.yml`

```yaml
name: Validate Telemetry Pipeline

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  validate-pipeline:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Checkout contract registry
        uses: actions/checkout@v4
        with:
          repository: myorg/telemetry-contracts
          path: ./contract-registry

      - name: Evaluate All Contracts
        uses: goedelsoup/waveform/actions/evaluate-contracts@v1
        with:
          contracts: './contract-registry/contracts/**/*.yaml'
          config: './collector-config.yaml'
          mode: 'pipeline'
          junit-output: 'test-results.xml'
          lcov-output: 'coverage.info'
          summary-output: 'summary.txt'
          verbose: 'true'
```

## Best Practices for Cross-Repository Setup

1. **Version Coordination**: Use semantic versioning for contracts and coordinate releases
2. **Artifact Management**: Set appropriate retention periods for contract artifacts
3. **Access Control**: Ensure proper permissions for cross-repository artifact access
4. **Monitoring**: Track contract compliance metrics across all services
5. **Rollback Strategy**: Maintain ability to rollback to previous contract versions
6. **Documentation**: Keep clear documentation of contract dependencies and versions
