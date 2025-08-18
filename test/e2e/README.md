# End-to-End Tests

This directory contains comprehensive end-to-end tests for the OpenTelemetry Contract Testing Framework. These tests validate the complete workflow from configuration loading to test execution and report generation.

## Test Coverage

### 1. Command Line Interface Tests (`TestEndToEnd_CommandLineInterface`)

Tests the full CLI workflow with various configuration scenarios:

- **ValidConfigurationAndContract**: Tests successful execution with valid configuration and contract files
- **MissingConfigFile**: Tests error handling when configuration file doesn't exist
- **InvalidYAML**: Tests error handling for malformed YAML configuration files
- **InvalidConfigurationStructure**: Tests validation of configuration structure (missing service section)
- **InvalidPipelineReferences**: Tests validation of pipeline component references

### 2. Report Generation Tests (`TestEndToEnd_ReportGeneration`)

Tests report generation functionality with configuration loading:

- **JUnitReport**: Tests JUnit XML report generation
- **LCOVReport**: Tests LCOV coverage report generation
- **SummaryReport**: Tests summary text report generation

### 3. Test Mode Tests (`TestEndToEnd_TestModes`)

Tests different testing modes with configuration:

- **PipelineMode**: Tests pipeline mode execution
- **ProcessorMode**: Tests processor mode execution

### 4. Help and Version Tests (`TestEndToEnd_HelpAndVersion`)

Tests CLI help and error handling:

- **Help**: Tests help command output
- **InvalidFlag**: Tests error handling for invalid command line flags

## Running the Tests

### Run All End-to-End Tests

```bash
just test-e2e
```

### Run Specific Test Categories

```bash
# Run only CLI tests
go test -v ./test/e2e/... -run TestEndToEnd_CommandLineInterface

# Run only report generation tests
go test -v ./test/e2e/... -run TestEndToEnd_ReportGeneration

# Run only test mode tests
go test -v ./test/e2e/... -run TestEndToEnd_TestModes
```

### Run with Verbose Output

```bash
go test -v ./test/e2e/... -count=1
```

## Test Structure

Each test follows this pattern:

1. **Setup**: Create temporary test directory and files
2. **Execution**: Run the waveform binary with specific arguments
3. **Validation**: Verify expected output and behavior
4. **Cleanup**: Temporary files are automatically cleaned up

## Test Data

The tests create various configuration and contract files:

### Valid Configuration Example
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024

exporters:
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
```

### Valid Contract Example
```yaml
publisher: "test-service"
pipeline: "traces"
version: "1.0"

inputs:
  traces:
    - span_name: "test_operation"
      service_name: "test-service"
      attributes:
        test.key: "test_value"

filters:
  - field: "span.service.name"
    operator: "equals"
    value: "test-service"

matchers:
  traces:
    - span_name: "test_operation"
      attributes:
        test.key: "test_value"
```

## Expected Behaviors

### Success Cases
- Configuration files are loaded and validated successfully
- Contracts are processed and tests are executed
- Reports are generated in the specified formats
- Exit code is 0 for successful runs

### Error Cases
- Missing configuration files result in appropriate error messages
- Invalid YAML files result in parsing error messages
- Invalid configuration structure results in validation error messages
- Invalid pipeline references result in specific error messages
- Exit code is non-zero for failed runs

## Integration with CI/CD

These end-to-end tests are designed to run in CI/CD pipelines and validate:

1. **Binary Build**: The waveform binary can be built successfully
2. **Configuration Loading**: Configuration files are loaded and validated correctly
3. **Contract Processing**: Contract files are processed and tests are executed
4. **Report Generation**: All report formats are generated correctly
5. **Error Handling**: Error conditions are handled gracefully with appropriate messages

## Performance

The end-to-end tests are designed to be fast and reliable:

- **Build Time**: ~1-2 seconds per test
- **Execution Time**: ~10-15ms per test execution
- **Total Suite Time**: ~5-6 seconds for all tests
- **Resource Usage**: Minimal, uses temporary directories that are automatically cleaned up

## Debugging

If tests fail, check:

1. **Build Issues**: Ensure the waveform binary can be built
2. **File Permissions**: Ensure temporary files can be created and written
3. **Output Differences**: Compare expected vs actual output for help text changes
4. **Configuration Format**: Verify configuration and contract file formats match expected schemas
