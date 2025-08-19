// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package collector

import (
	"testing"

	"github.com/goedelsoup/waveform/internal/harness"
	"go.uber.org/zap"
)

func TestService_Lifecycle(t *testing.T) {
	// Create a simple collector configuration
	config := harness.CollectorConfig{
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
				"timeout":         "1s",
				"send_batch_size": 1024,
			},
		},
		Exporters: map[string]interface{}{
			"logging": map[string]interface{}{
				"loglevel": "debug",
			},
		},
	}

	// Create logger
	logger := zap.NewNop()

	// Create service
	service := NewService(config, logger)

	// Test start
	if err := service.Start(); err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Test data processing
	testData := map[string]interface{}{
		"test": "data",
	}

	output, err := service.ProcessData(testData)
	if err != nil {
		t.Fatalf("Failed to process data: %v", err)
	}

	// Verify output
	if output == nil {
		t.Error("Expected non-nil output")
	}

	// Test stop
	if err := service.Stop(); err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}
}

func TestService_ComponentInitialization(t *testing.T) {
	// Create a configuration with various processor types
	config := harness.CollectorConfig{
		Receivers: map[string]interface{}{
			"otlp": map[string]interface{}{},
		},
		Processors: map[string]interface{}{
			"transform": map[string]interface{}{
				"traces": map[string]interface{}{
					"span": map[string]interface{}{
						"name": map[string]interface{}{
							"from_attributes": []string{"http.method", "http.route"},
						},
					},
				},
			},
			"attributes": map[string]interface{}{
				"actions": []map[string]interface{}{
					{
						"key":    "environment",
						"value":  "production",
						"action": "insert",
					},
				},
			},
			"filter": map[string]interface{}{
				"spans": map[string]interface{}{
					"include": map[string]interface{}{
						"match_type": "regexp",
						"attributes": []map[string]interface{}{
							{
								"key":   "http.status_code",
								"value": "4..|5..",
							},
						},
					},
				},
			},
		},
		Exporters: map[string]interface{}{
			"logging": map[string]interface{}{},
		},
	}

	logger := zap.NewNop()
	service := NewService(config, logger)

	// Test initialization
	if err := service.initializeComponents(); err != nil {
		t.Fatalf("Failed to initialize components: %v", err)
	}

	// Verify components were created
	if len(service.receivers) != 1 {
		t.Errorf("Expected 1 receiver, got %d", len(service.receivers))
	}

	if len(service.processors) != 3 {
		t.Errorf("Expected 3 processors, got %d", len(service.processors))
	}

	if len(service.exporters) != 1 {
		t.Errorf("Expected 1 exporter, got %d", len(service.exporters))
	}
}

func TestService_UnknownProcessor(t *testing.T) {
	// Test with an unknown processor type
	config := harness.CollectorConfig{
		Processors: map[string]interface{}{
			"unknown_processor": map[string]interface{}{
				"some_config": "value",
			},
		},
	}

	logger := zap.NewNop()
	service := NewService(config, logger)

	// This should not fail, but should create a mock processor
	if err := service.initializeComponents(); err != nil {
		t.Fatalf("Failed to initialize components with unknown processor: %v", err)
	}

	if len(service.processors) != 1 {
		t.Errorf("Expected 1 processor, got %d", len(service.processors))
	}
}
