<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io> -->

# Waveform

A standalone Go testing framework that applies contract testing principles to OpenTelemetry pipelines.

The framework allows telemetry publishers to define YAML contracts specifying their expectations, then validates that collector pipelines transform data correctly.

## What is Contract Testing?

Contract testing ensures that your OpenTelemetry collector pipelines correctly transform telemetry data according to your specifications. Instead of manually testing data transformations, you define contracts that specify:

- **Input data**: What telemetry data you're sending
- **Expected transformations**: How the data should be modified by your pipeline
- **Validation rules**: When and how to verify the transformations

This approach catches pipeline configuration errors early and ensures consistent data quality across your observability stack.

## Features

- **YAML Contract Definition**: Simple, declarative contracts for your telemetry expectations
- **Multiple Signal Types**: Test traces, metrics, and logs
- **Flexible Validation**: Conditional testing based on data characteristics
- **Two Testing Modes**: Test full collector pipelines or individual processors
- **Standard Reports**: JUnit XML and LCOV format outputs for CI/CD integration
- **Go Testing Integration**: Seamless integration with Go's testing ecosystem

## Quick Start

### 1. Install

```bash
go get github.com/goedelsoup/waveform
```

### 2. Define Your First Contract

Create a contract file `contracts/auth-service.yaml`:

```yaml
publisher: "auth-service"
pipeline: "http-trace"
version: "1.0"

inputs:
  traces:
    - span_name: "http_request"
      service_name: "auth-service"
      attributes:
        http.method: "POST"
        http.url: "https://auth.example.com/login"
        http.status_code: 200

filters:
  - field: "span.service.name"
    operator: "equals"
    value: "auth-service"

matchers:
  traces:
    - span_name: "http_request"
      attributes:
        http.method: "POST"
        normalized.method: "post"  # Expect method to be normalized to lowercase
        "!http.status_code": null  # Expect status_code to be removed
```

### 3. Run Your Tests

```bash
# Test against a collector configuration
waveform --contracts "./contracts/**/*.yaml" --config "./collector-config.yaml"

# Test individual processors
waveform --contracts "./contracts/**/*.yaml" --mode processor

# Generate test reports
waveform --contracts "./contracts/**/*.yaml" \
  --junit-output results.xml \
  --lcov-output coverage.info
```

## Contract Schema

### Basic Structure

```yaml
publisher: "service-name"          # Required: Your service identifier
pipeline: "pipeline-identifier"    # Required: Pipeline being tested
version: "1.0"                    # Required: Contract version
description: "Optional description"

inputs:                           # Required: Input data specification
  traces: [...]
  metrics: [...]
  logs: [...]

filters:                         # Optional: When to run this contract
  - field: "field.path"
    operator: "equals|not_equals|matches|exists|not_exists|greater_than|less_than"
    value: "expected_value"

matchers:                        # Required: Expected transformations
  traces: [...]
  metrics: [...]
  logs: [...]
```

### Input Examples

#### Trace Inputs

```yaml
inputs:
  traces:
    - span_name: "http_request"
      service_name: "my-service"
      attributes:
        http.method: "GET"
        http.url: "/api/users"
        user.id: "12345"
```

#### Metric Inputs

```yaml
inputs:
  metrics:
    - name: "request_count"
      type: "counter"
      value: 42
      labels:
        endpoint: "/api/users"
        method: "GET"
```

#### Log Inputs

```yaml
inputs:
  logs:
    - body: "User login successful"
      severity: "INFO"
      attributes:
        user.id: "12345"
        session.id: "abc123"
```

### Filter Operators

- `equals`: Exact string/number match
- `not_equals`: Inverse of equals
- `matches`: Regex pattern match
- `exists`: Field must exist
- `not_exists`: Field must not exist
- `greater_than`: Numeric comparison
- `less_than`: Numeric comparison

### Matcher Features

- **Negation**: Prefix field names with `!` to indicate they should not exist
- **Partial Matching**: Only specified fields are validated
- **Transformation Validation**: Verify that fields are transformed as expected

## CLI Usage

### Basic Commands

```bash
# Test all contracts in a directory
waveform --contracts "./contracts/**/*.yaml"

# Test with collector configuration
waveform --contracts "./contracts/**/*.yaml" --config collector.yaml

# Test individual processors
waveform --contracts "./contracts/**/*.yaml" --mode processor

# Generate test reports
waveform --contracts "./contracts/**/*.yaml" \
  --junit-output test-results.xml \
  --lcov-output coverage.info \
  --summary-output summary.txt
```

### Command Line Options

```bash
waveform [flags]

Flags:
  -c, --contracts strings     Contract file paths or glob patterns (required)
  -m, --mode string          Test mode: pipeline or processor (default "pipeline")
  -f, --config string        Collector configuration file path
  -j, --junit-output string  JUnit XML output file path
  -l, --lcov-output string   LCOV output file path
  -s, --summary-output string Summary output file path
  -v, --verbose              Enable verbose logging
```

## Integration with Go Tests

```go
package main

import (
    "testing"
    "github.com/goedelsoup/waveform/pkg/testing"
)

func TestAuthService(t *testing.T) {
    // Load contracts
    contracts, errors := testing.LoadContracts([]string{"./contracts/auth-service.yaml"})
    if len(errors) > 0 {
        t.Fatalf("Failed to load contracts: %v", errors)
    }

    // Run tests
    results, err := testing.RunContractTests(
        contracts,
        harness.TestModePipeline,
        config,
        nil,
    )
    if err != nil {
        t.Fatalf("Failed to run tests: %v", err)
    }

    // Check results
    if results.FailedTests > 0 {
        t.Errorf("Tests failed: %d/%d", results.FailedTests, results.TotalTests)
    }
}
```

## Examples

See the `examples/contracts/` directory for complete examples:

- `auth-service/http-trace.yaml`: HTTP trace validation
- `metrics-service/counter-metric.yaml`: Counter metric validation
- `logging-service/application-log.yaml`: Application log validation

## Best Practices

### Contract Design

1. **Be Specific**: Define precise input data that represents real scenarios
2. **Use Filters**: Apply filters to ensure contracts only run when relevant
3. **Test Transformations**: Focus on validating the actual transformations
4. **Version Contracts**: Use semantic versioning for contract evolution

### Test Organization

1. **Group by Publisher**: Organize contracts by service/publisher
2. **Use Descriptive Names**: Clear pipeline and contract names
3. **Include Examples**: Provide realistic input data
4. **Document Expectations**: Use descriptions to explain contract purpose

### Pipeline Testing

1. **Start Simple**: Begin with basic transformations
2. **Incremental Complexity**: Add complexity gradually
3. **Test Edge Cases**: Include boundary conditions
4. **Performance Considerations**: Use realistic data volumes

## Error Handling

The framework provides detailed error reporting:

- **Contract Loading Errors**: YAML parsing and validation errors
- **Test Execution Errors**: Collector startup and data processing errors
- **Validation Errors**: Detailed diff information for failed contracts
- **Configuration Errors**: Collector configuration issues

## Installation

### Binary Downloads

Download the latest release for your platform from the [GitHub releases page](https://github.com/goedelsoup/waveform/releases).

### Homebrew (macOS)

```bash
brew install goedelsoup/tap/waveform
```

### Docker

```bash
docker pull goedelsoup/waveform:latest
docker run --rm goedelsoup/waveform --help
```

### From Source

```bash
git clone https://github.com/goedelsoup/waveform.git
cd waveform
go install ./cmd/waveform
```

## Support

For questions and support:

- Create an issue on GitHub
- Check the examples directory
- Review the test files for usage patterns

## License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

## Development

For development information, see [.github/DEVELOPMENT.md](.github/DEVELOPMENT.md).

## Releasing

For information about the release process, see [RELEASING.md](RELEASING.md).
