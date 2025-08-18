// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEnd_ConfigurationLoading(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a valid collector configuration
	collectorConfig := `
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
	configPath := filepath.Join(tmpDir, "collector-config.yaml")
	err := os.WriteFile(configPath, []byte(collectorConfig), 0644)
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
	contractPath := filepath.Join(tmpDir, "test-contract.yaml")
	err = os.WriteFile(contractPath, []byte(contractData), 0644)
	require.NoError(t, err)

	// Test: Configuration loading with valid files
	t.Run("ValidConfiguration", func(t *testing.T) {
		// Set up command line arguments
		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", configPath,
			"--verbose",
		}

		// Capture stdout/stderr
		originalStdout := os.Stdout
		originalStderr := os.Stderr
		defer func() {
			os.Stdout = originalStdout
			os.Stderr = originalStderr
		}()

		// Run the command
		err := runCommand()
		assert.NoError(t, err, "Command should succeed with valid configuration")
	})

	// Test: Configuration loading with missing config file
	t.Run("MissingConfigFile", func(t *testing.T) {
		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", "/nonexistent/config.yaml",
			"--verbose",
		}

		err := runCommand()
		assert.Error(t, err, "Command should fail with missing config file")
		assert.Contains(t, err.Error(), "configuration file does not exist")
	})

	// Test: Configuration loading with invalid YAML
	t.Run("InvalidYAML", func(t *testing.T) {
		invalidConfig := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
    invalid: [yaml: structure
`
		invalidConfigPath := filepath.Join(tmpDir, "invalid-config.yaml")
		err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
		require.NoError(t, err)

		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", invalidConfigPath,
			"--verbose",
		}

		err = runCommand()
		assert.Error(t, err, "Command should fail with invalid YAML")
		assert.Contains(t, err.Error(), "invalid YAML format")
	})

	// Test: Configuration loading with invalid configuration structure
	t.Run("InvalidConfigurationStructure", func(t *testing.T) {
		invalidStructureConfig := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"

# Missing service section
`
		invalidStructurePath := filepath.Join(tmpDir, "invalid-structure.yaml")
		err := os.WriteFile(invalidStructurePath, []byte(invalidStructureConfig), 0644)
		require.NoError(t, err)

		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", invalidStructurePath,
			"--verbose",
		}

		err = runCommand()
		assert.Error(t, err, "Command should fail with invalid configuration structure")
		assert.Contains(t, err.Error(), "service section is required")
	})

	// Test: Configuration loading with invalid pipeline references
	t.Run("InvalidPipelineReferences", func(t *testing.T) {
		invalidPipelineConfig := `
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
      receivers: [nonexistent]  # Invalid receiver reference
      processors: [batch]
      exporters: [logging]
`
		invalidPipelinePath := filepath.Join(tmpDir, "invalid-pipeline.yaml")
		err := os.WriteFile(invalidPipelinePath, []byte(invalidPipelineConfig), 0644)
		require.NoError(t, err)

		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", invalidPipelinePath,
			"--verbose",
		}

		err = runCommand()
		assert.Error(t, err, "Command should fail with invalid pipeline references")
		assert.Contains(t, err.Error(), "receiver 'nonexistent' not found")
	})
}

func TestEndToEnd_MultipleConfigurationFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first configuration file
	config1 := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"

processors:
  batch:
    timeout: 1s

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
`

	// Create second configuration file
	config2 := `
exporters:
  logging:
    loglevel: debug

  otlp:
    endpoint: "http://localhost:4318"
`

	configPath1 := filepath.Join(tmpDir, "config1.yaml")
	configPath2 := filepath.Join(tmpDir, "config2.yaml")

	err := os.WriteFile(configPath1, []byte(config1), 0644)
	require.NoError(t, err)
	err = os.WriteFile(configPath2, []byte(config2), 0644)
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
	contractPath := filepath.Join(tmpDir, "test-contract.yaml")
	err = os.WriteFile(contractPath, []byte(contractData), 0644)
	require.NoError(t, err)

	// Test: Multiple configuration files (this would require LoadFromPaths support)
	// For now, we'll test that the framework can handle the merged configuration
	// by creating a single merged file
	mergedConfig := `
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
  otlp:
    endpoint: "http://localhost:4318"

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging, otlp]
`
	mergedConfigPath := filepath.Join(tmpDir, "merged-config.yaml")
	err = os.WriteFile(mergedConfigPath, []byte(mergedConfig), 0644)
	require.NoError(t, err)

	os.Args = []string{
		"waveform",
		"--contracts", contractPath,
		"--config", mergedConfigPath,
		"--verbose",
	}

	err = runCommand()
	assert.NoError(t, err, "Command should succeed with merged configuration")
}

func TestEndToEnd_ReportGeneration(t *testing.T) {
	tmpDir := t.TempDir()

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
	configPath := filepath.Join(tmpDir, "collector-config.yaml")
	err := os.WriteFile(configPath, []byte(configData), 0644)
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
	contractPath := filepath.Join(tmpDir, "test-contract.yaml")
	err = os.WriteFile(contractPath, []byte(contractData), 0644)
	require.NoError(t, err)

	// Test: Report generation with configuration
	t.Run("JUnitReport", func(t *testing.T) {
		junitOutput := filepath.Join(tmpDir, "test-results.xml")
		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", configPath,
			"--junit-output", junitOutput,
			"--verbose",
		}

		err := runCommand()
		assert.NoError(t, err, "Command should succeed with JUnit report generation")

		// Verify report file was created
		_, err = os.Stat(junitOutput)
		assert.NoError(t, err, "JUnit report file should be created")
	})

	t.Run("LCOVReport", func(t *testing.T) {
		lcovOutput := filepath.Join(tmpDir, "coverage.info")
		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", configPath,
			"--lcov-output", lcovOutput,
			"--verbose",
		}

		err := runCommand()
		assert.NoError(t, err, "Command should succeed with LCOV report generation")

		// Verify report file was created
		_, err = os.Stat(lcovOutput)
		assert.NoError(t, err, "LCOV report file should be created")
	})

	t.Run("SummaryReport", func(t *testing.T) {
		summaryOutput := filepath.Join(tmpDir, "summary.txt")
		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", configPath,
			"--summary-output", summaryOutput,
			"--verbose",
		}

		err := runCommand()
		assert.NoError(t, err, "Command should succeed with summary report generation")

		// Verify report file was created
		_, err = os.Stat(summaryOutput)
		assert.NoError(t, err, "Summary report file should be created")
	})
}

func TestEndToEnd_TestModes(t *testing.T) {
	tmpDir := t.TempDir()

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
	configPath := filepath.Join(tmpDir, "collector-config.yaml")
	err := os.WriteFile(configPath, []byte(configData), 0644)
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
	contractPath := filepath.Join(tmpDir, "test-contract.yaml")
	err = os.WriteFile(contractPath, []byte(contractData), 0644)
	require.NoError(t, err)

	// Test: Pipeline mode
	t.Run("PipelineMode", func(t *testing.T) {
		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", configPath,
			"--mode", "pipeline",
			"--verbose",
		}

		err := runCommand()
		assert.NoError(t, err, "Command should succeed in pipeline mode")
	})

	// Test: Processor mode
	t.Run("ProcessorMode", func(t *testing.T) {
		os.Args = []string{
			"waveform",
			"--contracts", contractPath,
			"--config", configPath,
			"--mode", "processor",
			"--verbose",
		}

		err := runCommand()
		assert.NoError(t, err, "Command should succeed in processor mode")
	})
}

// runCommand is a helper function to run the main command for testing
func runCommand() error {
	// Reset global variables to avoid test interference
	contractPaths = []string{}
	testMode = "pipeline"
	configPath = ""
	junitOutput = ""
	lcovOutput = ""
	summaryOutput = ""
	verbose = false

	// Create a new root command for each test
	rootCmd := &cobra.Command{
		Use:   "waveform",
		Short: "OpenTelemetry Contract Testing Framework",
		Long: `A standalone Go testing framework that applies contract testing principles
to OpenTelemetry pipelines. The framework allows telemetry publishers to define
YAML contracts specifying their expectations, then validates that collector
pipelines transform data correctly.`,
		RunE: runTests,
	}

	// Add flags
	rootCmd.Flags().StringSliceVarP(&contractPaths, "contracts", "c", []string{}, "Contract file paths or glob patterns")
	rootCmd.Flags().StringVarP(&testMode, "mode", "m", "pipeline", "Test mode: pipeline or processor")
	rootCmd.Flags().StringVarP(&configPath, "config", "f", "", "Collector configuration file path")
	rootCmd.Flags().StringVarP(&junitOutput, "junit-output", "j", "", "JUnit XML output file path")
	rootCmd.Flags().StringVarP(&lcovOutput, "lcov-output", "l", "", "LCOV output file path")
	rootCmd.Flags().StringVarP(&summaryOutput, "summary-output", "s", "", "Summary output file path")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	// Mark required flags
	if err := rootCmd.MarkFlagRequired("contracts"); err != nil {
		return fmt.Errorf("failed to mark contracts flag as required: %w", err)
	}

	// Parse the command line arguments
	rootCmd.SetArgs(os.Args[1:])
	return rootCmd.Execute()
}
