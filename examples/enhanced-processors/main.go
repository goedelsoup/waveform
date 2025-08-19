// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package main

import (
	"fmt"
	"log"

	"github.com/goedelsoup/waveform/internal/collector"
	"github.com/goedelsoup/waveform/internal/contract"
	"github.com/goedelsoup/waveform/internal/generator"
	"github.com/goedelsoup/waveform/internal/harness"
	"github.com/goedelsoup/waveform/internal/matcher"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func main() {
	// Create a logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	logger.Info("Starting Enhanced Processors Example")

	// Create a sample contract that expects transformed data
	contractDef := &contract.Contract{
		Publisher: "example-service",
		Pipeline:  "traces",
		Version:   "1.0",
		Inputs: contract.Inputs{
			Traces: []contract.TraceInput{
				{
					SpanName: "http-request",
					Attributes: map[string]interface{}{
						"http.method":      "GET",
						"http.route":       "/api/users",
						"http.status_code": 200,
					},
				},
			},
		},
		Matchers: contract.Matchers{
			Traces: []contract.TraceMatcher{
				{
					SpanName: "GET /api/users", // Expect transformed span name
					Attributes: map[string]interface{}{
						"environment":  "production",
						"service.name": "waveform",
						"http.method":  "GET",
						"http.route":   "/api/users",
					},
				},
			},
		},
	}

	// Create collector configuration with enhanced processors
	collectorConfig := harness.CollectorConfig{
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
			"attributes": map[string]interface{}{
				"actions": []map[string]interface{}{
					{
						"key":    "environment",
						"value":  "production",
						"action": "insert",
					},
					{
						"key":    "service.name",
						"value":  "waveform",
						"action": "insert",
					},
				},
			},
			"transform": map[string]interface{}{
				"traces": map[string]interface{}{
					"span": map[string]interface{}{
						"name": map[string]interface{}{
							"from_attributes": []string{"http.method", "http.route"},
							"separator":       " ",
						},
					},
				},
			},
		},
		Exporters: map[string]interface{}{
			"logging": map[string]interface{}{
				"loglevel": "debug",
			},
		},
	}

	// Create components
	gen := generator.NewGenerator()
	mat := matcher.NewMatcher()

	// Generate test data
	logger.Info("Generating test data")
	inputData := gen.GenerateFromContract(contractDef)

	// Log original data
	logger.Info("Original input data",
		zap.Int("traces", inputData.Traces.ResourceSpans().Len()),
		zap.Int("metrics", inputData.Metrics.ResourceMetrics().Len()),
		zap.Int("logs", inputData.Logs.ResourceLogs().Len()))

	// Print original span details
	if inputData.Traces.ResourceSpans().Len() > 0 {
		span := inputData.Traces.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
		logger.Info("Original span",
			zap.String("name", span.Name()),
			zap.Any("attributes", spanAttributesToMap(span.Attributes())))
	}

	// Create collector service
	logger.Info("Creating collector service with enhanced processors")
	collectorService := collector.NewService(collectorConfig, logger)

	// Start the collector service
	logger.Info("Starting collector service")
	if err := collectorService.Start(); err != nil {
		logger.Fatal("Failed to start collector service", zap.Error(err))
	}
	defer func() {
		if err := collectorService.Stop(); err != nil {
			logger.Error("Failed to stop collector service", zap.Error(err))
		}
	}()

	// Process data through collector
	logger.Info("Processing data through enhanced processors")
	output, err := collectorService.ProcessData(inputData)
	if err != nil {
		logger.Fatal("Failed to process data", zap.Error(err))
	}

	// Convert output back to OpenTelemetryData
	outputData, ok := output.(contract.OpenTelemetryData)
	if !ok {
		// If conversion fails, create empty data
		outputData = contract.OpenTelemetryData{
			Traces:  ptrace.NewTraces(),
			Metrics: pmetric.NewMetrics(),
			Logs:    plog.NewLogs(),
		}
	}

	// Log processed data
	logger.Info("Processed output data",
		zap.Int("traces", outputData.Traces.ResourceSpans().Len()),
		zap.Int("metrics", outputData.Metrics.ResourceMetrics().Len()),
		zap.Int("logs", outputData.Logs.ResourceLogs().Len()))

	// Print processed span details
	if outputData.Traces.ResourceSpans().Len() > 0 {
		span := outputData.Traces.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
		logger.Info("Processed span",
			zap.String("name", span.Name()),
			zap.Any("attributes", spanAttributesToMap(span.Attributes())))
	}

	// Validate the output
	logger.Info("Validating output against contract")
	validationResult := mat.Validate(contractDef, inputData, outputData)

	// Print results
	fmt.Println("\n=== Enhanced Processors Example Results ===")
	fmt.Printf("Input traces: %d\n", inputData.Traces.ResourceSpans().Len())
	fmt.Printf("Output traces: %d\n", outputData.Traces.ResourceSpans().Len())
	fmt.Printf("Validation passed: %t\n", validationResult.Valid)

	if !validationResult.Valid {
		fmt.Println("Validation errors:")
		for _, err := range validationResult.Errors {
			fmt.Printf("  - %s\n", err.Message)
		}
	} else {
		fmt.Println("✅ All validations passed! Enhanced processors are working correctly.")
	}

	// Demonstrate the harness integration
	logger.Info("Demonstrating harness integration with enhanced processors")
	demonstrateHarnessIntegration(logger, contractDef, collectorConfig, collectorService)
}

func demonstrateHarnessIntegration(logger *zap.Logger, contractDef *contract.Contract, config harness.CollectorConfig, collectorService *collector.Service) {
	// Create test harness
	harness := harness.NewTestHarness(harness.TestModePipeline, config)
	harness.SetLogger(logger)
	harness.SetCollectorService(collectorService)

	// Run the test
	logger.Info("Running test with enhanced processors")
	results := harness.RunTests([]*contract.Contract{contractDef})
	if len(results.Results) > 0 {
		result := results.Results[0]

		// Print results
		fmt.Println("\n=== Harness Integration Results ===")
		fmt.Printf("Test passed: %t\n", result.Valid)
		fmt.Printf("Duration: %v\n", result.Duration)
		fmt.Printf("Error count: %d\n", len(result.Errors))

		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for _, err := range result.Errors {
				fmt.Printf("  - %s\n", err)
			}
		} else {
			fmt.Println("✅ Harness integration successful!")
		}
	}
}

// Helper function to convert span attributes to a map for logging
func spanAttributesToMap(attributes pcommon.Map) map[string]string {
	result := make(map[string]string)
	attributes.Range(func(k string, v pcommon.Value) bool {
		result[k] = v.AsString()
		return true
	})
	return result
}
