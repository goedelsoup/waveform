// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// RunnerConfig represents the configuration for the Waveform runner
type RunnerConfig struct {
	// Runner settings
	Runner RunnerSettings `yaml:"runner" toml:"runner"`

	// Collector definitions for matching system
	Collectors map[string]CollectorDefinition `yaml:"collectors" toml:"collectors"`

	// Pipeline selectors for dynamic pipeline matching
	PipelineSelectors []PipelineSelector `yaml:"pipeline_selectors" toml:"pipeline_selectors"`

	// Global settings
	Global GlobalSettings `yaml:"global" toml:"global"`
}

// RunnerSettings contains runner-specific configuration
type RunnerSettings struct {
	// Logging configuration
	LogLevel string `yaml:"log_level" toml:"log_level" default:"info"`
	LogFormat string `yaml:"log_format" toml:"log_format" default:"json"`

	// Test execution settings
	Timeout Duration `yaml:"timeout" toml:"timeout" default:"30s"`
	Parallel int `yaml:"parallel" toml:"parallel" default:"1"`

	// Output settings
	Output OutputSettings `yaml:"output" toml:"output"`

	// Cache settings
	Cache CacheSettings `yaml:"cache" toml:"cache"`
}

// OutputSettings configures output behavior
type OutputSettings struct {
	// Report formats to generate
	Formats []string `yaml:"formats" toml:"formats" default:"[\"summary\"]"`

	// Output directory for reports
	Directory string `yaml:"directory" toml:"directory" default:"./waveform-reports"`

	// Whether to overwrite existing files
	Overwrite bool `yaml:"overwrite" toml:"overwrite" default:"false"`

	// Verbose output
	Verbose bool `yaml:"verbose" toml:"verbose" default:"false"`
}

// CacheSettings configures caching behavior
type CacheSettings struct {
	// Whether to enable caching
	Enabled bool `yaml:"enabled" toml:"enabled" default:"true"`

	// Cache directory
	Directory string `yaml:"directory" toml:"directory" default:"~/.cache/waveform"`

	// Cache TTL
	TTL Duration `yaml:"ttl" toml:"ttl" default:"1h"`

	// Maximum cache size in MB
	MaxSize int `yaml:"max_size" toml:"max_size" default:"100"`
}

// CollectorDefinition defines a collector configuration
type CollectorDefinition struct {
	// Collector name/identifier
	Name string `yaml:"name" toml:"name"`

	// Collector description
	Description string `yaml:"description" toml:"description"`

	// Collector configuration file path
	ConfigPath string `yaml:"config_path" toml:"config_path"`

	// Environment-specific settings
	Environment map[string]interface{} `yaml:"environment" toml:"environment"`

	// Tags for categorization
	Tags []string `yaml:"tags" toml:"tags"`

	// Pipeline configurations
	Pipelines map[string]PipelineConfig `yaml:"pipelines" toml:"pipelines"`
}

// PipelineConfig defines pipeline-specific settings
type PipelineConfig struct {
	// Pipeline name
	Name string `yaml:"name" toml:"name"`

	// Pipeline description
	Description string `yaml:"description" toml:"description"`

	// Signal types this pipeline handles
	Signals []string `yaml:"signals" toml:"signals"`

	// Pipeline selectors for matching
	Selectors []PipelineSelector `yaml:"selectors" toml:"selectors"`

	// Environment-specific overrides
	Environment map[string]interface{} `yaml:"environment" toml:"environment"`
}

// PipelineSelector defines criteria for pipeline matching
type PipelineSelector struct {
	// Field to match against
	Field string `yaml:"field" toml:"field"`

	// Operator for comparison
	Operator string `yaml:"operator" toml:"operator"`

	// Value to match against
	Value interface{} `yaml:"value" toml:"value"`

	// Priority for this selector (higher = more specific)
	Priority int `yaml:"priority" toml:"priority" default:"0"`
}

// GlobalSettings contains global configuration
type GlobalSettings struct {
	// Default environment
	Environment string `yaml:"environment" toml:"environment" default:"development"`

	// Default timeout for operations
	DefaultTimeout Duration `yaml:"default_timeout" toml:"default_timeout" default:"30s"`

	// Whether to fail fast on errors
	FailFast bool `yaml:"fail_fast" toml:"fail_fast" default:"false"`

	// Retry settings
	Retry RetrySettings `yaml:"retry" toml:"retry"`
}

// RetrySettings configures retry behavior
type RetrySettings struct {
	// Maximum number of retries
	MaxAttempts int `yaml:"max_attempts" toml:"max_attempts" default:"3"`

	// Initial backoff duration
	InitialBackoff Duration `yaml:"initial_backoff" toml:"initial_backoff" default:"1s"`

	// Maximum backoff duration
	MaxBackoff Duration `yaml:"max_backoff" toml:"max_backoff" default:"30s"`

	// Backoff multiplier
	BackoffMultiplier float64 `yaml:"backoff_multiplier" toml:"backoff_multiplier" default:"2.0"`
}

// Duration represents a duration string that can be parsed
type Duration string

// RunnerConfigLoader handles loading runner configuration files
type RunnerConfigLoader struct{}

// NewRunnerConfigLoader creates a new runner configuration loader
func NewRunnerConfigLoader() *RunnerConfigLoader {
	return &RunnerConfigLoader{}
}

// LoadConfig loads the runner configuration from the appropriate location
func (l *RunnerConfigLoader) LoadConfig() (*RunnerConfig, error) {
	// Try to find configuration file in order of preference
	configPaths := l.getConfigPaths()

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return l.LoadFromFile(path)
		}
	}

	// Return default configuration if no file found
	return l.getDefaultConfig(), nil
}

// LoadFromFile loads configuration from a specific file
func (l *RunnerConfigLoader) LoadFromFile(path string) (*RunnerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", path, err)
	}

	// Determine file format based on extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return l.parseYAML(data)
	case ".toml":
		return l.parseTOML(data)
	default:
		return nil, fmt.Errorf("unsupported configuration file format: %s", ext)
	}
}

// getConfigPaths returns the list of configuration file paths to check
func (l *RunnerConfigLoader) getConfigPaths() []string {
	var paths []string

	// Current directory (highest priority)
	cwd, err := os.Getwd()
	if err == nil {
		paths = append(paths,
			filepath.Join(cwd, ".waveform.yaml"),
			filepath.Join(cwd, "waveform.yaml"),
			filepath.Join(cwd, ".waveform.toml"),
			filepath.Join(cwd, "waveform.toml"),
		)
	}

	// XDG config directory
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		paths = append(paths,
			filepath.Join(xdgConfig, "waveform", "config.yaml"),
			filepath.Join(xdgConfig, "waveform", "config.toml"),
		)
	} else {
		// Default XDG config location
		homeDir, err := os.UserHomeDir()
		if err == nil {
			paths = append(paths,
				filepath.Join(homeDir, ".config", "waveform", "config.yaml"),
				filepath.Join(homeDir, ".config", "waveform", "config.toml"),
			)
		}
	}

	// Legacy home directory (lowest priority)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".waveform.yaml"),
			filepath.Join(homeDir, "waveform.yaml"),
			filepath.Join(homeDir, ".waveform.toml"),
			filepath.Join(homeDir, "waveform.toml"),
		)
	}

	return paths
}

// parseYAML parses YAML configuration data
func (l *RunnerConfigLoader) parseYAML(data []byte) (*RunnerConfig, error) {
	var config RunnerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML configuration: %w", err)
	}
	return &config, nil
}

// parseTOML parses TOML configuration data
func (l *RunnerConfigLoader) parseTOML(data []byte) (*RunnerConfig, error) {
	var config RunnerConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML configuration: %w", err)
	}
	return &config, nil
}

// getDefaultConfig returns the default configuration
func (l *RunnerConfigLoader) getDefaultConfig() *RunnerConfig {
	return &RunnerConfig{
		Runner: RunnerSettings{
			LogLevel:  "info",
			LogFormat: "json",
			Timeout:   "30s",
			Parallel:  1,
			Output: OutputSettings{
				Formats:   []string{"summary"},
				Directory: "./waveform-reports",
				Overwrite: false,
				Verbose:   false,
			},
			Cache: CacheSettings{
				Enabled:  true,
				Directory: "~/.cache/waveform",
				TTL:      "1h",
				MaxSize:  100,
			},
		},
		Collectors: make(map[string]CollectorDefinition),
		PipelineSelectors: []PipelineSelector{},
		Global: GlobalSettings{
			Environment:    "development",
			DefaultTimeout: "30s",
			FailFast:       false,
			Retry: RetrySettings{
				MaxAttempts:      3,
				InitialBackoff:   "1s",
				MaxBackoff:       "30s",
				BackoffMultiplier: 2.0,
			},
		},
	}
}

// SaveConfig saves the configuration to a file
func (l *RunnerConfigLoader) SaveConfig(config *RunnerConfig, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Determine format based on extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML configuration: %w", err)
		}
		return os.WriteFile(path, data, 0644)
	case ".toml":
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create TOML file: %w", err)
		}
		defer file.Close()
		return toml.NewEncoder(file).Encode(config)
	default:
		return fmt.Errorf("unsupported configuration file format: %s", ext)
	}
}

// GetCollectorConfig returns the collector configuration for a given name
func (l *RunnerConfigLoader) GetCollectorConfig(name string) (*CollectorDefinition, error) {
	config, err := l.LoadConfig()
	if err != nil {
		return nil, err
	}

	collector, exists := config.Collectors[name]
	if !exists {
		return nil, fmt.Errorf("collector '%s' not found in configuration", name)
	}

	return &collector, nil
}

// GetPipelineSelectors returns pipeline selectors for a given pipeline
func (l *RunnerConfigLoader) GetPipelineSelectors(pipelineName string) []PipelineSelector {
	config, err := l.LoadConfig()
	if err != nil {
		return nil
	}

	var selectors []PipelineSelector
	
	// Check global pipeline selectors
	for _, selector := range config.PipelineSelectors {
		if selector.Field == "pipeline.name" && selector.Operator == "equals" && selector.Value == pipelineName {
			selectors = append(selectors, selector)
		}
	}

	// Check collector-specific selectors
	for _, collector := range config.Collectors {
		if pipeline, exists := collector.Pipelines[pipelineName]; exists {
			selectors = append(selectors, pipeline.Selectors...)
		}
	}

	return selectors
}
