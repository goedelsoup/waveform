// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goedelsoup/waveform/internal/harness"
	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing configuration files
type Loader struct{}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromFile loads a configuration from a single file
func (l *Loader) LoadFromFile(path string) (*harness.CollectorConfig, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file does not exist: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", path, err)
	}

	// Parse configuration
	config, err := l.parseConfig(data, path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file %s: %w", path, err)
	}

	return config, nil
}

// LoadFromPaths loads configuration from multiple files or glob patterns
func (l *Loader) LoadFromPaths(paths []string) (*harness.CollectorConfig, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no configuration paths provided")
	}

	// If only one path is provided, load it directly
	if len(paths) == 1 {
		return l.LoadFromFile(paths[0])
	}

	// For multiple paths, merge configurations
	var mergedConfig *harness.CollectorConfig
	var errors []string

	for _, path := range paths {
		config, err := l.LoadFromFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", path, err))
			continue
		}

		if mergedConfig == nil {
			mergedConfig = config
		} else {
			mergedConfig = l.mergeConfigs(mergedConfig, config)
		}
	}

	if mergedConfig == nil {
		return nil, fmt.Errorf("failed to load any configuration files: %s", strings.Join(errors, "; "))
	}

	if len(errors) > 0 {
		// Log warnings for failed files but continue with merged config
		fmt.Fprintf(os.Stderr, "Warning: Some configuration files failed to load: %s\n", strings.Join(errors, "; "))
	}

	return mergedConfig, nil
}

// parseConfig parses YAML configuration data
func (l *Loader) parseConfig(data []byte, source string) (*harness.CollectorConfig, error) {
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("invalid YAML format: %w", err)
	}

	config := &harness.CollectorConfig{
		Receivers:  make(map[string]interface{}),
		Processors: make(map[string]interface{}),
		Exporters:  make(map[string]interface{}),
		Service:    make(map[string]interface{}),
	}

	// Extract receivers
	if receivers, ok := rawConfig["receivers"]; ok {
		if receiversMap, ok := receivers.(map[string]interface{}); ok {
			config.Receivers = receiversMap
		} else {
			return nil, fmt.Errorf("receivers section must be a map in %s", source)
		}
	}

	// Extract processors
	if processors, ok := rawConfig["processors"]; ok {
		if processorsMap, ok := processors.(map[string]interface{}); ok {
			config.Processors = processorsMap
		} else {
			return nil, fmt.Errorf("processors section must be a map in %s", source)
		}
	}

	// Extract exporters
	if exporters, ok := rawConfig["exporters"]; ok {
		if exportersMap, ok := exporters.(map[string]interface{}); ok {
			config.Exporters = exportersMap
		} else {
			return nil, fmt.Errorf("exporters section must be a map in %s", source)
		}
	}

	// Extract service
	if service, ok := rawConfig["service"]; ok {
		if serviceMap, ok := service.(map[string]interface{}); ok {
			config.Service = serviceMap
		} else {
			return nil, fmt.Errorf("service section must be a map in %s", source)
		}
	}

	return config, nil
}

// mergeConfigs merges two configurations, with the second config taking precedence
func (l *Loader) mergeConfigs(base, override *harness.CollectorConfig) *harness.CollectorConfig {
	merged := &harness.CollectorConfig{
		Receivers:  make(map[string]interface{}),
		Processors: make(map[string]interface{}),
		Exporters:  make(map[string]interface{}),
		Service:    make(map[string]interface{}),
	}

	// Merge receivers
	for k, v := range base.Receivers {
		merged.Receivers[k] = v
	}
	for k, v := range override.Receivers {
		merged.Receivers[k] = v
	}

	// Merge processors
	for k, v := range base.Processors {
		merged.Processors[k] = v
	}
	for k, v := range override.Processors {
		merged.Processors[k] = v
	}

	// Merge exporters
	for k, v := range base.Exporters {
		merged.Exporters[k] = v
	}
	for k, v := range override.Exporters {
		merged.Exporters[k] = v
	}

	// Merge service
	for k, v := range base.Service {
		merged.Service[k] = v
	}
	for k, v := range override.Service {
		merged.Service[k] = v
	}

	return merged
}

// ValidateConfig validates the configuration structure
func (l *Loader) ValidateConfig(config *harness.CollectorConfig) error {
	var errors []string

	// Check if service section exists and has pipelines
	if len(config.Service) == 0 {
		errors = append(errors, "service section is required")
	} else {
		// Check for pipelines in service
		if service, ok := config.Service["pipelines"]; ok {
			if pipelines, ok := service.(map[string]interface{}); ok {
				for pipelineName, pipeline := range pipelines {
					if err := l.validatePipeline(pipeline, pipelineName, config); err != nil {
						errors = append(errors, fmt.Sprintf("pipeline %s: %v", pipelineName, err))
					}
				}
			} else {
				errors = append(errors, "service.pipelines must be a map")
			}
		} else {
			errors = append(errors, "service.pipelines section is required")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validatePipeline validates a single pipeline configuration
func (l *Loader) validatePipeline(pipeline interface{}, pipelineName string, config *harness.CollectorConfig) error {
	pipelineMap, ok := pipeline.(map[string]interface{})
	if !ok {
		return fmt.Errorf("pipeline must be a map")
	}

	// Check for required pipeline components
	receivers, _ := pipelineMap["receivers"].([]interface{})
	processors, _ := pipelineMap["processors"].([]interface{})
	exporters, _ := pipelineMap["exporters"].([]interface{})

	// Validate receivers
	for _, receiver := range receivers {
		if receiverName, ok := receiver.(string); ok {
			if _, exists := config.Receivers[receiverName]; !exists {
				return fmt.Errorf("receiver '%s' not found in receivers section", receiverName)
			}
		}
	}

	// Validate processors
	for _, processor := range processors {
		if processorName, ok := processor.(string); ok {
			if _, exists := config.Processors[processorName]; !exists {
				return fmt.Errorf("processor '%s' not found in processors section", processorName)
			}
		}
	}

	// Validate exporters
	for _, exporter := range exporters {
		if exporterName, ok := exporter.(string); ok {
			if _, exists := config.Exporters[exporterName]; !exists {
				return fmt.Errorf("exporter '%s' not found in exporters section", exporterName)
			}
		}
	}

	return nil
}

// SaveConfig saves a configuration to a file
func (l *Loader) SaveConfig(config *harness.CollectorConfig, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file %s: %w", path, err)
	}

	return nil
}

// LoadFromReader loads configuration from an io.Reader
func (l *Loader) LoadFromReader(reader io.Reader, source string) (*harness.CollectorConfig, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration from %s: %w", source, err)
	}

	return l.parseConfig(data, source)
}
