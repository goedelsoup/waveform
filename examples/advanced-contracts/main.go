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

	logger.Info("Starting Advanced Contracts Example")

	// Create an advanced contract with sophisticated validation rules
	advancedContract := createAdvancedContract()

	// Create components
	gen := generator.NewGenerator()
	mat := matcher.NewMatcher()

	// Generate test data
	logger.Info("Generating test data for advanced contract")
	inputData := gen.GenerateFromContract(advancedContract)

	// Create collector configuration
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
						"key":    "payment.status",
						"value":  "completed",
						"action": "insert",
					},
					{
						"key":    "customer.tier",
						"value":  "premium",
						"action": "insert",
					},
					{
						"key":    "payment.total_cents",
						"value":  29999,
						"action": "insert",
					},
				},
			},
			"transform": map[string]interface{}{
				"traces": map[string]interface{}{
					"span": map[string]interface{}{
						"name": map[string]interface{}{
							"from_attributes": []string{"http.method", "payment.type"},
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

	// Create collector service
	logger.Info("Creating collector service with advanced processing")
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
	logger.Info("Processing data through advanced pipeline")
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

	// Validate against advanced contract
	logger.Info("Validating output against advanced contract")
	validationResult := mat.Validate(advancedContract, inputData, outputData)

	// Print detailed results
	printAdvancedValidationResults(validationResult)

	// Demonstrate the harness integration
	logger.Info("Demonstrating harness integration with advanced contract")
	demonstrateAdvancedHarnessIntegration(logger, advancedContract, collectorConfig, collectorService)
}

func createAdvancedContract() *contract.Contract {
	return &contract.Contract{
		Publisher:   "payment-service",
		Version:     "2.0.0",
		Description: "Advanced payment service contract with sophisticated validation",
		PipelineSelectors: &contract.PipelineSelectors{
			Selectors: []contract.PipelineSelector{
				{
					Field:    "type",
					Operator: contract.PipelineSelectorOperatorEquals,
					Value:    "trace",
				},
				{
					Field:    "tags.service",
					Operator: contract.PipelineSelectorOperatorMatches,
					Value:    "payment.*",
				},
			},
			Priority: 10,
		},
		Inputs: contract.Inputs{
			Traces: []contract.TraceInput{
				{
					SpanName:    "process_payment",
					ServiceName: "payment-service",
					Attributes: map[string]interface{}{
						"payment.method":   "credit_card",
						"payment.amount":   299.99,
						"http.method":      "POST",
						"http.status_code": 200,
					},
				},
			},
		},
		ValidationRules: []contract.ValidationRule{
			// Range validation for payment amounts
			{
				Field:    "span.attributes.payment.amount",
				Operator: contract.FilterOperatorInRange,
				Range: &contract.ValueRange{
					Min:       0.01,
					Max:       10000.00,
					Inclusive: true,
				},
				Description: "Payment amount must be between $0.01 and $10,000",
				Severity:    contract.SeverityError,
			},
			// Pattern validation for payment status
			{
				Field:       "span.attributes.payment.status",
				Operator:    contract.FilterOperatorOneOf,
				Values:      []interface{}{"completed", "pending", "failed", "cancelled"},
				Description: "Payment status must be valid",
				Severity:    contract.SeverityError,
			},
			// Conditional validation: Premium customers get special treatment
			{
				Field:    "span.attributes.processing_priority",
				Operator: contract.FilterOperatorEquals,
				Value:    "high",
				Condition: &contract.ConditionalRule{
					If: &contract.ValidationRule{
						Field:    "span.attributes.customer.tier",
						Operator: contract.FilterOperatorEquals,
						Value:    "premium",
					},
				},
				Description: "Premium customers should get high priority processing",
				Severity:    contract.SeverityWarning,
			},
		},
		Matchers: contract.Matchers{
			Traces: []contract.TraceMatcher{
				{
					SpanName:    "POST payment",
					ServiceName: "payment-service",
					Attributes: map[string]interface{}{
						"payment.method":      "credit_card",
						"payment.status":      "completed",
						"payment.total_cents": 29999,
						"customer.tier":       "premium",
						"http.status_code":    200,
					},
					ValidationRules: []contract.ValidationRule{
						{
							Field:       "span.attributes.payment.status",
							Operator:    contract.FilterOperatorOneOf,
							Values:      []interface{}{"completed", "authorized", "captured"},
							Description: "Payment must be in a valid final state",
						},
						{
							Field:       "span.attributes.payment.total_cents",
							Operator:    contract.FilterOperatorExists,
							Description: "Payment amount should be converted to cents",
						},
					},
					Count: &contract.CountMatcher{
						Expected: 1,
						Operator: contract.FilterOperatorEquals,
					},
					Duration: &contract.DurationMatcher{
						Min:       "50ms",
						Max:       "5s",
						Expected:  "200ms",
						Tolerance: "100ms",
					},
					StatusCode: &contract.StatusCodeMatcher{
						Class:      "2xx",
						NotAllowed: []int{400, 401, 403, 500, 502, 503, 504},
					},
				},
			},
		},
		TimeWindows: []contract.TimeWindow{
			{
				Aggregation:      "p95",
				Duration:         "5m",
				ExpectedBehavior: "payment_processing_sla",
			},
		},
		Schema: &contract.ContractSchema{
			Version:        "2.0",
			RequiredFields: []string{"publisher", "version", "inputs", "matchers"},
			FieldTypes: map[string]string{
				"publisher": "string",
				"version":   "string",
			},
			ValidationRules: []contract.SchemaValidationRule{
				{
					Field:       "publisher",
					Type:        "string",
					Required:    true,
					Pattern:     "^[a-z][a-z0-9-]*[a-z0-9]$",
					Description: "Publisher must be lowercase with hyphens",
				},
			},
		},
	}
}

func printAdvancedValidationResults(result contract.ValidationResult) {
	fmt.Println("\n=== Advanced Contract Validation Results ===")
	fmt.Printf("Overall validation: %t\n", result.Valid)

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for i, err := range result.Errors {
			fmt.Printf("  %d. [%s] %s\n", i+1, err.Type, err.Message)
			if err.Field != "" {
				fmt.Printf("     Field: %s\n", err.Field)
			}
			if err.Expected != nil {
				fmt.Printf("     Expected: %v, Actual: %v\n", err.Expected, err.Actual)
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for i, warning := range result.Warnings {
			fmt.Printf("  %d. %s\n", i+1, warning)
		}
	}

	if result.Valid {
		fmt.Println("✅ All advanced validations passed!")
	} else {
		fmt.Println("❌ Some advanced validations failed!")
	}
}

func demonstrateAdvancedHarnessIntegration(logger *zap.Logger, contractDef *contract.Contract, config harness.CollectorConfig, collectorService *collector.Service) {
	// Create test harness
	testHarness := harness.NewTestHarness(harness.TestModePipeline, config)
	testHarness.SetLogger(logger)
	testHarness.SetCollectorService(collectorService)

	// Run the test
	logger.Info("Running advanced contract test")
	results := testHarness.RunTests([]*contract.Contract{contractDef})

	if len(results.Results) > 0 {
		result := results.Results[0]

		// Print results
		fmt.Println("\n=== Advanced Harness Integration Results ===")
		fmt.Printf("Test passed: %t\n", result.Valid)
		fmt.Printf("Duration: %v\n", result.Duration)
		fmt.Printf("Error count: %d\n", len(result.Errors))

		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for i, err := range result.Errors {
				fmt.Printf("  %d. %s\n", i+1, err)
			}
		} else {
			fmt.Println("✅ Advanced harness integration successful!")
		}

		// Print advanced features summary
		fmt.Println("\n=== Advanced Features Demonstrated ===")
		fmt.Println("• Sophisticated validation rules with ranges and patterns")
		fmt.Println("• Conditional validation logic (if-then-else)")
		fmt.Println("• Multi-level validation severity (error, warning, info)")
		fmt.Println("• Advanced matchers with count, duration, and status code validation")
		fmt.Println("• Temporal validation rules for performance monitoring")
		fmt.Println("• Schema validation for contract structure")
		fmt.Println("• Pipeline selectors for dynamic matching")
		fmt.Println("• Cross-signal validation capabilities")
	}
}
