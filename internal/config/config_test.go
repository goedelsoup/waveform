// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goedelsoup/waveform/internal/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadFromFile(t *testing.T) {
	loader := NewLoader()

	// Create a temporary test file
	testConfig := `
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

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Load configuration
	config, err := loader.LoadFromFile(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify configuration structure
	assert.NotNil(t, config.Receivers)
	assert.NotNil(t, config.Processors)
	assert.NotNil(t, config.Exporters)
	assert.NotNil(t, config.Service)

	// Verify specific values
	assert.Contains(t, config.Receivers, "otlp")
	assert.Contains(t, config.Processors, "batch")
	assert.Contains(t, config.Exporters, "logging")

	// Verify service pipelines
	service, ok := config.Service["pipelines"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, service, "traces")
}

func TestLoader_LoadFromFile_NotExists(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFromFile("/nonexistent/file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestLoader_LoadFromFile_InvalidYAML(t *testing.T) {
	loader := NewLoader()

	// Create a temporary test file with invalid YAML
	testConfig := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
    invalid: [yaml: structure
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yaml")
	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Load configuration
	_, err = loader.LoadFromFile(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid YAML format")
}

func TestLoader_LoadFromPaths_SingleFile(t *testing.T) {
	loader := NewLoader()

	// Create a temporary test file
	testConfig := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [logging]
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Load configuration
	config, err := loader.LoadFromPaths([]string{configPath})
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Contains(t, config.Receivers, "otlp")
}

func TestLoader_LoadFromPaths_MultipleFiles(t *testing.T) {
	loader := NewLoader()

	// Create first config file
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

	// Create second config file
	config2 := `
exporters:
  logging:
    loglevel: debug

  otlp:
    endpoint: "http://localhost:4318"
`

	tmpDir := t.TempDir()
	configPath1 := filepath.Join(tmpDir, "config1.yaml")
	configPath2 := filepath.Join(tmpDir, "config2.yaml")

	err := os.WriteFile(configPath1, []byte(config1), 0644)
	require.NoError(t, err)
	err = os.WriteFile(configPath2, []byte(config2), 0644)
	require.NoError(t, err)

	// Load configuration
	config, err := loader.LoadFromPaths([]string{configPath1, configPath2})
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify merged configuration
	assert.Contains(t, config.Receivers, "otlp")
	assert.Contains(t, config.Processors, "batch")
	assert.Contains(t, config.Exporters, "logging")
	assert.Contains(t, config.Exporters, "otlp")
}

func TestLoader_ValidateConfig_Valid(t *testing.T) {
	loader := NewLoader()

	config := &harness.CollectorConfig{
		Receivers: map[string]interface{}{
			"otlp": map[string]interface{}{
				"protocols": map[string]interface{}{
					"grpc": map[string]interface{}{
						"endpoint": "0.0.0.0:4317",
					},
				},
			},
		},
		Processors: map[string]interface{}{
			"batch": map[string]interface{}{
				"timeout": "1s",
			},
		},
		Exporters: map[string]interface{}{
			"logging": map[string]interface{}{
				"loglevel": "debug",
			},
		},
		Service: map[string]interface{}{
			"pipelines": map[string]interface{}{
				"traces": map[string]interface{}{
					"receivers":  []interface{}{"otlp"},
					"processors": []interface{}{"batch"},
					"exporters":  []interface{}{"logging"},
				},
			},
		},
	}

	err := loader.ValidateConfig(config)
	assert.NoError(t, err)
}

func TestLoader_ValidateConfig_MissingService(t *testing.T) {
	loader := NewLoader()

	config := &harness.CollectorConfig{
		Receivers:  make(map[string]interface{}),
		Processors: make(map[string]interface{}),
		Exporters:  make(map[string]interface{}),
		Service:    make(map[string]interface{}),
	}

	err := loader.ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service section is required")
}

func TestLoader_ValidateConfig_InvalidReceiver(t *testing.T) {
	loader := NewLoader()

	config := &harness.CollectorConfig{
		Receivers: map[string]interface{}{
			"otlp": map[string]interface{}{
				"protocols": map[string]interface{}{
					"grpc": map[string]interface{}{
						"endpoint": "0.0.0.0:4317",
					},
				},
			},
		},
		Processors: make(map[string]interface{}),
		Exporters:  make(map[string]interface{}),
		Service: map[string]interface{}{
			"pipelines": map[string]interface{}{
				"traces": map[string]interface{}{
					"receivers":  []interface{}{"nonexistent"},
					"processors": []interface{}{},
					"exporters":  []interface{}{"logging"},
				},
			},
		},
	}

	err := loader.ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "receiver 'nonexistent' not found")
}

func TestLoader_SaveConfig(t *testing.T) {
	loader := NewLoader()

	config := &harness.CollectorConfig{
		Receivers: map[string]interface{}{
			"otlp": map[string]interface{}{
				"protocols": map[string]interface{}{
					"grpc": map[string]interface{}{
						"endpoint": "0.0.0.0:4317",
					},
				},
			},
		},
		Service: map[string]interface{}{
			"pipelines": map[string]interface{}{
				"traces": map[string]interface{}{
					"receivers":  []interface{}{"otlp"},
					"processors": []interface{}{},
					"exporters":  []interface{}{"logging"},
				},
			},
		},
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved-config.yaml")

	// Save configuration
	err := loader.SaveConfig(config, configPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify the saved configuration
	loadedConfig, err := loader.LoadFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, config.Receivers, loadedConfig.Receivers)
	assert.Equal(t, config.Service, loadedConfig.Service)
}

func TestLoader_LoadFromReader(t *testing.T) {
	loader := NewLoader()

	testConfig := `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [logging]
`

	// Create a reader from string
	reader := strings.NewReader(testConfig)

	// Load configuration
	config, err := loader.LoadFromReader(reader, "test-source")
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Contains(t, config.Receivers, "otlp")
}
