// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerConfigLoader_LoadConfig_Default(t *testing.T) {
	loader := NewRunnerConfigLoader()
	config, err := loader.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify default values
	assert.Equal(t, "info", config.Runner.LogLevel)
	assert.Equal(t, "json", config.Runner.LogFormat)
	assert.Equal(t, "30s", string(config.Runner.Timeout))
	assert.Equal(t, 1, config.Runner.Parallel)
	assert.Equal(t, "./waveform-reports", config.Runner.Output.Directory)
	assert.Equal(t, []string{"summary"}, config.Runner.Output.Formats)
	assert.True(t, config.Runner.Cache.Enabled)
	assert.Equal(t, "development", config.Global.Environment)
}

func TestRunnerConfigLoader_LoadFromFile_YAML(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a temporary YAML config file
	testConfig := `
runner:
  log_level: debug
  log_format: console
  timeout: 60s
  parallel: 4
  output:
    formats: ["junit", "lcov", "summary"]
    directory: "/tmp/reports"
    overwrite: true
    verbose: true
  cache:
    enabled: false
    directory: "/tmp/cache"
    ttl: 2h
    max_size: 200

collectors:
  production:
    name: "production-collector"
    description: "Production environment collector"
    config_path: "/etc/collector/production.yaml"
    tags: ["production", "main"]
    pipelines:
      traces:
        name: "traces-pipeline"
        description: "Trace processing pipeline"
        signals: ["traces"]
        selectors:
          - field: "type"
            operator: "equals"
            value: "trace"
            priority: 1

  staging:
    name: "staging-collector"
    description: "Staging environment collector"
    config_path: "/etc/collector/staging.yaml"
    tags: ["staging", "test"]
    pipelines:
      metrics:
        name: "metrics-pipeline"
        description: "Metrics processing pipeline"
        signals: ["metrics"]
        selectors:
          - field: "type"
            operator: "equals"
            value: "metric"
            priority: 2

pipeline_selectors:
  - field: "environment"
    operator: "equals"
    value: "production"
    priority: 10

global:
  environment: production
  default_timeout: 60s
  fail_fast: true
  retry:
    max_attempts: 5
    initial_backoff: 2s
    max_backoff: 60s
    backoff_multiplier: 1.5
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Load configuration
	config, err := loader.LoadFromFile(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify runner settings
	assert.Equal(t, "debug", config.Runner.LogLevel)
	assert.Equal(t, "console", config.Runner.LogFormat)
	assert.Equal(t, "60s", string(config.Runner.Timeout))
	assert.Equal(t, 4, config.Runner.Parallel)

	// Verify output settings
	assert.Equal(t, "/tmp/reports", config.Runner.Output.Directory)
	assert.Equal(t, []string{"junit", "lcov", "summary"}, config.Runner.Output.Formats)
	assert.True(t, config.Runner.Output.Overwrite)
	assert.True(t, config.Runner.Output.Verbose)

	// Verify cache settings
	assert.False(t, config.Runner.Cache.Enabled)
	assert.Equal(t, "/tmp/cache", config.Runner.Cache.Directory)
	assert.Equal(t, "2h", string(config.Runner.Cache.TTL))
	assert.Equal(t, 200, config.Runner.Cache.MaxSize)

	// Verify collectors
	assert.Len(t, config.Collectors, 2)
	
	prodCollector, exists := config.Collectors["production"]
	assert.True(t, exists)
	assert.Equal(t, "production-collector", prodCollector.Name)
	assert.Equal(t, "Production environment collector", prodCollector.Description)
	assert.Equal(t, "/etc/collector/production.yaml", prodCollector.ConfigPath)
	assert.Equal(t, []string{"production", "main"}, prodCollector.Tags)

	// Verify pipelines
	assert.Len(t, prodCollector.Pipelines, 1)
	tracesPipeline, exists := prodCollector.Pipelines["traces"]
	assert.True(t, exists)
	assert.Equal(t, "traces-pipeline", tracesPipeline.Name)
	assert.Equal(t, []string{"traces"}, tracesPipeline.Signals)
	assert.Len(t, tracesPipeline.Selectors, 1)
	assert.Equal(t, "type", tracesPipeline.Selectors[0].Field)
	assert.Equal(t, "equals", tracesPipeline.Selectors[0].Operator)
	assert.Equal(t, "trace", tracesPipeline.Selectors[0].Value)
	assert.Equal(t, 1, tracesPipeline.Selectors[0].Priority)

	// Verify global settings
	assert.Equal(t, "production", config.Global.Environment)
	assert.Equal(t, "60s", string(config.Global.DefaultTimeout))
	assert.True(t, config.Global.FailFast)
	assert.Equal(t, 5, config.Global.Retry.MaxAttempts)
	assert.Equal(t, "2s", string(config.Global.Retry.InitialBackoff))
	assert.Equal(t, "60s", string(config.Global.Retry.MaxBackoff))
	assert.Equal(t, 1.5, config.Global.Retry.BackoffMultiplier)

	// Verify pipeline selectors
	assert.Len(t, config.PipelineSelectors, 1)
	assert.Equal(t, "environment", config.PipelineSelectors[0].Field)
	assert.Equal(t, "equals", config.PipelineSelectors[0].Operator)
	assert.Equal(t, "production", config.PipelineSelectors[0].Value)
	assert.Equal(t, 10, config.PipelineSelectors[0].Priority)
}

func TestRunnerConfigLoader_LoadFromFile_TOML(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a temporary TOML config file
	testConfig := `
[runner]
log_level = "debug"
log_format = "console"
timeout = "60s"
parallel = 4

[runner.output]
formats = ["junit", "lcov", "summary"]
directory = "/tmp/reports"
overwrite = true
verbose = true

[runner.cache]
enabled = false
directory = "/tmp/cache"
ttl = "2h"
max_size = 200

[collectors.production]
name = "production-collector"
description = "Production environment collector"
config_path = "/etc/collector/production.yaml"
tags = ["production", "main"]

[collectors.production.pipelines.traces]
name = "traces-pipeline"
description = "Trace processing pipeline"
signals = ["traces"]

[[collectors.production.pipelines.traces.selectors]]
field = "type"
operator = "equals"
value = "trace"
priority = 1

[collectors.staging]
name = "staging-collector"
description = "Staging environment collector"
config_path = "/etc/collector/staging.yaml"
tags = ["staging", "test"]

[collectors.staging.pipelines.metrics]
name = "metrics-pipeline"
description = "Metrics processing pipeline"
signals = ["metrics"]

[[collectors.staging.pipelines.metrics.selectors]]
field = "type"
operator = "equals"
value = "metric"
priority = 2

[[pipeline_selectors]]
field = "environment"
operator = "equals"
value = "production"
priority = 10

[global]
environment = "production"
default_timeout = "60s"
fail_fast = true

[global.retry]
max_attempts = 5
initial_backoff = "2s"
max_backoff = "60s"
backoff_multiplier = 1.5
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.toml")
	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Load configuration
	config, err := loader.LoadFromFile(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify runner settings
	assert.Equal(t, "debug", config.Runner.LogLevel)
	assert.Equal(t, "console", config.Runner.LogFormat)
	assert.Equal(t, "60s", string(config.Runner.Timeout))
	assert.Equal(t, 4, config.Runner.Parallel)

	// Verify output settings
	assert.Equal(t, "/tmp/reports", config.Runner.Output.Directory)
	assert.Equal(t, []string{"junit", "lcov", "summary"}, config.Runner.Output.Formats)
	assert.True(t, config.Runner.Output.Overwrite)
	assert.True(t, config.Runner.Output.Verbose)

	// Verify cache settings
	assert.False(t, config.Runner.Cache.Enabled)
	assert.Equal(t, "/tmp/cache", config.Runner.Cache.Directory)
	assert.Equal(t, "2h", string(config.Runner.Cache.TTL))
	assert.Equal(t, 200, config.Runner.Cache.MaxSize)

	// Verify collectors
	assert.Len(t, config.Collectors, 2)
	
	prodCollector, exists := config.Collectors["production"]
	assert.True(t, exists)
	assert.Equal(t, "production-collector", prodCollector.Name)
	assert.Equal(t, "Production environment collector", prodCollector.Description)
	assert.Equal(t, "/etc/collector/production.yaml", prodCollector.ConfigPath)
	assert.Equal(t, []string{"production", "main"}, prodCollector.Tags)

	// Verify global settings
	assert.Equal(t, "production", config.Global.Environment)
	assert.Equal(t, "60s", string(config.Global.DefaultTimeout))
	assert.True(t, config.Global.FailFast)
	assert.Equal(t, 5, config.Global.Retry.MaxAttempts)
}

func TestRunnerConfigLoader_LoadFromFile_InvalidFormat(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a temporary file with invalid extension
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")
	err := os.WriteFile(configPath, []byte(`{"test": "data"}`), 0644)
	require.NoError(t, err)

	// Try to load configuration
	_, err = loader.LoadFromFile(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported configuration file format")
}

func TestRunnerConfigLoader_LoadFromFile_InvalidYAML(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a temporary YAML file with invalid syntax
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(configPath, []byte(`invalid: yaml: [syntax`), 0644)
	require.NoError(t, err)

	// Try to load configuration
	_, err = loader.LoadFromFile(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML configuration")
}

func TestRunnerConfigLoader_LoadFromFile_InvalidTOML(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a temporary TOML file with invalid syntax
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.toml")
	err := os.WriteFile(configPath, []byte(`[invalid toml syntax`), 0644)
	require.NoError(t, err)

	// Try to load configuration
	_, err = loader.LoadFromFile(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse TOML configuration")
}

func TestRunnerConfigLoader_GetConfigPaths(t *testing.T) {
	loader := NewRunnerConfigLoader()
	paths := loader.getConfigPaths()

	// Should have multiple paths in order of priority
	assert.Greater(t, len(paths), 0)

	// Current directory paths should be first
	assert.Contains(t, paths[0], ".waveform.yaml")
	assert.Contains(t, paths[1], "waveform.yaml")
	assert.Contains(t, paths[2], ".waveform.toml")
	assert.Contains(t, paths[3], "waveform.toml")

	// Should include XDG config paths
	foundXDG := false
	for _, path := range paths {
		if strings.Contains(path, ".config/waveform") {
			foundXDG = true
			break
		}
	}
	assert.True(t, foundXDG, "Should include XDG config paths")
}

func TestRunnerConfigLoader_GetCollectorConfig(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a temporary config file with collectors
	testConfig := `
collectors:
  test-collector:
    name: "test-collector"
    description: "Test collector"
    config_path: "/tmp/test.yaml"
    tags: ["test"]
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "waveform.yaml")
	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Change to the temp directory to load the config
	originalCwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalCwd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Get collector config
	collector, err := loader.GetCollectorConfig("test-collector")
	require.NoError(t, err)
	assert.Equal(t, "test-collector", collector.Name)
	assert.Equal(t, "Test collector", collector.Description)
	assert.Equal(t, "/tmp/test.yaml", collector.ConfigPath)
	assert.Equal(t, []string{"test"}, collector.Tags)

	// Test non-existent collector
	_, err = loader.GetCollectorConfig("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collector 'non-existent' not found")
}

func TestRunnerConfigLoader_GetPipelineSelectors(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a temporary config file with pipeline selectors
	testConfig := `
pipeline_selectors:
  - field: "pipeline.name"
    operator: "equals"
    value: "test-pipeline"
    priority: 5

collectors:
  test-collector:
    name: "test-collector"
    pipelines:
      test-pipeline:
        name: "test-pipeline"
        selectors:
          - field: "type"
            operator: "equals"
            value: "trace"
            priority: 1
          - field: "environment"
            operator: "equals"
            value: "production"
            priority: 2
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "waveform.yaml")
	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Change to the temp directory to load the config
	originalCwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalCwd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Get pipeline selectors
	selectors := loader.GetPipelineSelectors("test-pipeline")
	assert.Len(t, selectors, 3)

	// Should include global selector
	foundGlobal := false
	for _, selector := range selectors {
		if selector.Field == "pipeline.name" && selector.Value == "test-pipeline" {
			foundGlobal = true
			assert.Equal(t, 5, selector.Priority)
			break
		}
	}
	assert.True(t, foundGlobal, "Should include global pipeline selector")

	// Should include collector-specific selectors
	foundType := false
	foundEnv := false
	for _, selector := range selectors {
		if selector.Field == "type" && selector.Value == "trace" {
			foundType = true
			assert.Equal(t, 1, selector.Priority)
		}
		if selector.Field == "environment" && selector.Value == "production" {
			foundEnv = true
			assert.Equal(t, 2, selector.Priority)
		}
	}
	assert.True(t, foundType, "Should include type selector")
	assert.True(t, foundEnv, "Should include environment selector")
}

func TestRunnerConfigLoader_SaveConfig_YAML(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a test configuration
	config := &RunnerConfig{
		Runner: RunnerSettings{
			LogLevel:  "debug",
			LogFormat: "console",
			Timeout:   "60s",
			Parallel:  4,
		},
		Collectors: map[string]CollectorDefinition{
			"test": {
				Name:        "test-collector",
				Description: "Test collector",
				ConfigPath:  "/tmp/test.yaml",
			},
		},
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-save.yaml")

	// Save configuration
	err := loader.SaveConfig(config, configPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify the saved configuration
	loadedConfig, err := loader.LoadFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "debug", loadedConfig.Runner.LogLevel)
	assert.Equal(t, "console", loadedConfig.Runner.LogFormat)
	assert.Equal(t, "60s", string(loadedConfig.Runner.Timeout))
	assert.Equal(t, 4, loadedConfig.Runner.Parallel)

	collector, exists := loadedConfig.Collectors["test"]
	assert.True(t, exists)
	assert.Equal(t, "test-collector", collector.Name)
	assert.Equal(t, "Test collector", collector.Description)
	assert.Equal(t, "/tmp/test.yaml", collector.ConfigPath)
}

func TestRunnerConfigLoader_SaveConfig_TOML(t *testing.T) {
	loader := NewRunnerConfigLoader()

	// Create a test configuration
	config := &RunnerConfig{
		Runner: RunnerSettings{
			LogLevel:  "debug",
			LogFormat: "console",
			Timeout:   "60s",
			Parallel:  4,
		},
		Collectors: map[string]CollectorDefinition{
			"test": {
				Name:        "test-collector",
				Description: "Test collector",
				ConfigPath:  "/tmp/test.yaml",
			},
		},
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-save.toml")

	// Save configuration
	err := loader.SaveConfig(config, configPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify the saved configuration
	loadedConfig, err := loader.LoadFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "debug", loadedConfig.Runner.LogLevel)
	assert.Equal(t, "console", loadedConfig.Runner.LogFormat)
	assert.Equal(t, "60s", string(loadedConfig.Runner.Timeout))
	assert.Equal(t, 4, loadedConfig.Runner.Parallel)

	collector, exists := loadedConfig.Collectors["test"]
	assert.True(t, exists)
	assert.Equal(t, "test-collector", collector.Name)
	assert.Equal(t, "Test collector", collector.Description)
	assert.Equal(t, "/tmp/test.yaml", collector.ConfigPath)
}
