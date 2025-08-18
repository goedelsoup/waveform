// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEnd_CommandLineInterface tests the full CLI workflow
func TestEndToEnd_CommandLineInterface(t *testing.T) {
	// Build the binary for testing
	buildDir := t.TempDir()
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(buildDir, "waveform"), "./cmd/waveform")
	buildCmd.Dir = "../../"
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build waveform binary")

	binaryPath := filepath.Join(buildDir, "waveform")

	// Create test directory
	testDir := t.TempDir()

	t.Run("ValidConfigurationAndContract", func(t *testing.T) {
		// Create a valid collector configuration
		configData := `
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
`
		configPath := filepath.Join(testDir, "collector-config.yaml")
		err := os.WriteFile(configPath, []byte(configData), 0644)
		require.NoError(t, err)

		// Create a valid contract
		contractData := `
publisher: "test-service"
pipeline: "traces"
version: "1.0"
description: "End-to-end test contract"

inputs:
  traces:
    - span_name: "test_operation"
      service_name: "test-service"
      attributes:
        test.key: "test_value"
        http.method: "GET"

filters:
  - field: "span.service.name"
    operator: "equals"
    value: "test-service"

matchers:
  traces:
    - span_name: "test_operation"
      attributes:
        test.key: "test_value"
        http.method: "GET"
`
		contractPath := filepath.Join(testDir, "test-contract.yaml")
		err = os.WriteFile(contractPath, []byte(contractData), 0644)
		require.NoError(t, err)

		// Run the command
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", configPath, "--verbose")
		output, err := cmd.CombinedOutput()

		// The command should succeed
		assert.NoError(t, err, "Command should succeed with valid configuration and contract")
		assert.Contains(t, string(output), "Collector configuration loaded successfully")
		assert.Contains(t, string(output), "Contracts loaded successfully")
	})

	t.Run("MissingConfigFile", func(t *testing.T) {
		// Create a valid contract
		contractData := `
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
`
		contractPath := filepath.Join(testDir, "test-contract.yaml")
		err := os.WriteFile(contractPath, []byte(contractData), 0644)
		require.NoError(t, err)

		// Run the command with non-existent config file
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", "/nonexistent/config.yaml", "--verbose")
		output, err := cmd.CombinedOutput()

		// The command should fail with appropriate error message
		assert.Error(t, err, "Command should fail with missing config file")
		assert.Contains(t, string(output), "configuration file does not exist")
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		// Create invalid YAML configuration
		invalidConfig := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
    invalid: [yaml: structure
`
		invalidConfigPath := filepath.Join(testDir, "invalid-config.yaml")
		err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
		require.NoError(t, err)

		// Create a valid contract
		contractData := `
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
`
		contractPath := filepath.Join(testDir, "test-contract.yaml")
		err = os.WriteFile(contractPath, []byte(contractData), 0644)
		require.NoError(t, err)

		// Run the command
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", invalidConfigPath, "--verbose")
		output, err := cmd.CombinedOutput()

		// The command should fail with YAML parsing error
		assert.Error(t, err, "Command should fail with invalid YAML")
		assert.Contains(t, string(output), "invalid YAML format")
	})

	t.Run("InvalidConfigurationStructure", func(t *testing.T) {
		// Create configuration with missing service section
		invalidStructureConfig := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"

# Missing service section
`
		invalidStructurePath := filepath.Join(testDir, "invalid-structure.yaml")
		err := os.WriteFile(invalidStructurePath, []byte(invalidStructureConfig), 0644)
		require.NoError(t, err)

		// Create a valid contract
		contractData := `
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
`
		contractPath := filepath.Join(testDir, "test-contract.yaml")
		err = os.WriteFile(contractPath, []byte(contractData), 0644)
		require.NoError(t, err)

		// Run the command
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", invalidStructurePath, "--verbose")
		output, err := cmd.CombinedOutput()

		// The command should fail with validation error
		assert.Error(t, err, "Command should fail with invalid configuration structure")
		assert.Contains(t, string(output), "service section is required")
	})
}

// TestEndToEnd_ReportGeneration tests report generation with configuration
func TestEndToEnd_ReportGeneration(t *testing.T) {
	// Build the binary for testing
	buildDir := t.TempDir()
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(buildDir, "waveform"), "./cmd/waveform")
	buildCmd.Dir = "../../"
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build waveform binary")

	binaryPath := filepath.Join(buildDir, "waveform")
	testDir := t.TempDir()

	// Create configuration and contract files
	configData := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"

processors:
  batch:
    timeout: 1s

exporters:
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
`
	configPath := filepath.Join(testDir, "collector-config.yaml")
	err = os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	contractData := `
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
`
	contractPath := filepath.Join(testDir, "test-contract.yaml")
	err = os.WriteFile(contractPath, []byte(contractData), 0644)
	require.NoError(t, err)

	t.Run("JUnitReport", func(t *testing.T) {
		junitOutput := filepath.Join(testDir, "test-results.xml")
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", configPath, "--junit-output", junitOutput, "--verbose")
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Command should succeed with JUnit report generation")
		assert.Contains(t, string(output), "Generating JUnit XML report")

		// Verify report file was created
		_, err = os.Stat(junitOutput)
		assert.NoError(t, err, "JUnit report file should be created")
	})

	t.Run("LCOVReport", func(t *testing.T) {
		lcovOutput := filepath.Join(testDir, "coverage.info")
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", configPath, "--lcov-output", lcovOutput, "--verbose")
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Command should succeed with LCOV report generation")
		assert.Contains(t, string(output), "Generating LCOV report")

		// Verify report file was created
		_, err = os.Stat(lcovOutput)
		assert.NoError(t, err, "LCOV report file should be created")
	})

	t.Run("SummaryReport", func(t *testing.T) {
		summaryOutput := filepath.Join(testDir, "summary.txt")
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", configPath, "--summary-output", summaryOutput, "--verbose")
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Command should succeed with summary report generation")
		assert.Contains(t, string(output), "Generating summary report")

		// Verify report file was created
		_, err = os.Stat(summaryOutput)
		assert.NoError(t, err, "Summary report file should be created")
	})
}

// TestEndToEnd_TestModes tests different test modes with configuration
func TestEndToEnd_TestModes(t *testing.T) {
	// Build the binary for testing
	buildDir := t.TempDir()
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(buildDir, "waveform"), "./cmd/waveform")
	buildCmd.Dir = "../../"
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build waveform binary")

	binaryPath := filepath.Join(buildDir, "waveform")
	testDir := t.TempDir()

	// Create configuration and contract files
	configData := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"

processors:
  batch:
    timeout: 1s

exporters:
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
`
	configPath := filepath.Join(testDir, "collector-config.yaml")
	err = os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	contractData := `
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
`
	contractPath := filepath.Join(testDir, "test-contract.yaml")
	err = os.WriteFile(contractPath, []byte(contractData), 0644)
	require.NoError(t, err)

	t.Run("PipelineMode", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", configPath, "--mode", "pipeline", "--verbose")
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Command should succeed in pipeline mode")
		assert.Contains(t, string(output), "Running tests")
		assert.Contains(t, string(output), "mode")
	})

	t.Run("ProcessorMode", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--contracts", contractPath, "--config", configPath, "--mode", "processor", "--verbose")
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Command should succeed in processor mode")
		assert.Contains(t, string(output), "Running tests")
		assert.Contains(t, string(output), "mode")
	})
}

// TestEndToEnd_HelpAndVersion tests help and version commands
func TestEndToEnd_HelpAndVersion(t *testing.T) {
	// Build the binary for testing
	buildDir := t.TempDir()
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(buildDir, "waveform"), "./cmd/waveform")
	buildCmd.Dir = "../../"
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build waveform binary")

	binaryPath := filepath.Join(buildDir, "waveform")

	t.Run("Help", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Help command should succeed")
		assert.Contains(t, string(output), "OpenTelemetry pipelines")
		assert.Contains(t, string(output), "--contracts")
		assert.Contains(t, string(output), "--config")
		assert.Contains(t, string(output), "waveform")
	})

	t.Run("InvalidFlag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--invalid-flag")
		output, err := cmd.CombinedOutput()

		assert.Error(t, err, "Command should fail with invalid flag")
		assert.Contains(t, string(output), "unknown flag")
	})
}
