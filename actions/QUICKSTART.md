# Quick Start Guide

This guide will help you get started with Waveform GitHub Actions in just a few minutes.

## Prerequisites

- A GitHub repository with OpenTelemetry contracts
- Go 1.21+ (the actions will install this automatically)
- Basic understanding of YAML and GitHub Actions

## Step 1: Create Your First Contract

Create a directory structure for your contracts:

```bash
mkdir -p contracts/my-service
```

Create your first contract file `contracts/my-service/http-trace.yaml`:

```yaml
publisher: "my-service"
pipeline: "http-trace"
version: "1.0"
description: "HTTP request/response trace validation"

inputs:
  traces:
    - span_name: "http_request"
      service_name: "my-service"
      attributes:
        http.method: "GET"
        http.url: "/api/users"
        http.status_code: 200

filters:
  - field: "span.service.name"
    operator: "equals"
    value: "my-service"

matchers:
  traces:
    - span_name: "http_request"
      attributes:
        http.method: "GET"
        normalized.method: "get"  # Expect method to be normalized
        "!http.status_code": null  # Expect status_code to be removed
```

## Step 2: Set Up Application Repository Workflow

Create `.github/workflows/register-contracts.yml` in your application repository:

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

## Step 3: Set Up Pipeline Repository Workflow

Create `.github/workflows/validate-pipeline.yml` in your pipeline repository:

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

      - name: Download Service Contracts
        uses: actions/download-artifact@v3
        with:
          name: registered-contracts
          path: ./service-contracts

      - name: Evaluate Contracts
        uses: goedelsoup/waveform/actions/evaluate-contracts@v1
        with:
          contracts: './service-contracts/**/*.yaml'
          config: './collector-config.yaml'
          mode: 'pipeline'
          verbose: 'true'
```

## Step 4: Create Collector Configuration

Create `collector-config.yaml` in your pipeline repository:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  transform:
    traces:
      spans:
        - from_attributes: ["http.method"]
          to_attributes: ["normalized.method"]
          action: insert
        - from_attributes: ["http.status_code"]
          action: delete

exporters:
  logging:
    verbosity: detailed

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [transform]
      exporters: [logging]
```

## Step 5: Test Your Setup

1. **Push to a feature branch**: Create a pull request to test contract validation
2. **Check the workflow**: Verify that the contract validation passes
3. **Merge to main**: The contracts will be registered as artifacts
4. **Test pipeline validation**: The pipeline repository will download and test the contracts

## Step 6: Monitor Results

Check the workflow results:

- **Pull Request**: Contract validation results in the PR checks
- **Main Branch**: Contract registration artifacts in the workflow run
- **Pipeline Repository**: Test results and coverage reports

## Common Issues and Solutions

### Issue: "Action not found"
**Solution**: The actions need to be published to the GitHub Marketplace first. For now, you can use the actions directly from the repository.

### Issue: "No contracts found"
**Solution**: Check that your contract files have the `.yaml` extension and are in the specified path.

### Issue: "Contract validation failed"
**Solution**: Check the YAML syntax and ensure all required fields are present in your contracts.

### Issue: "Pipeline validation failed"
**Solution**: Verify that your collector configuration matches the contract expectations.

## Next Steps

1. **Add more contracts**: Create contracts for different telemetry types (metrics, logs)
2. **Enhance validation**: Add more complex filters and matchers
3. **Set up monitoring**: Track contract compliance over time
4. **Integrate with other services**: Add contracts for additional microservices

## Getting Help

- Check the [main documentation](../README.md)
- Review the [examples](../examples/) directory
- Create an issue on the GitHub repository
- Check the action-specific documentation in each action directory
