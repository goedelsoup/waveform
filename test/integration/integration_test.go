// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/goedelsoup/waveform/internal/contract"
	"github.com/goedelsoup/waveform/internal/harness"
	"github.com/goedelsoup/waveform/internal/matcher"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// TestIntegration_BasicFunctionality tests basic framework functionality
func TestIntegration_BasicFunctionality(t *testing.T) {
	t.Run("ContractValidation", func(t *testing.T) {
		// Test contract structure validation
		contractDef := &contract.Contract{
			Publisher: "test-service",
			Pipeline:  "traces",
			Version:   "1.0",
			Inputs: contract.Inputs{
				Traces: []contract.TraceInput{
					{
						SpanName:    "test_operation",
						ServiceName: "test-service",
						Attributes: map[string]interface{}{
							"test.key": "test_value",
						},
					},
				},
			},
			Matchers: contract.Matchers{
				Traces: []contract.TraceMatcher{
					{
						SpanName: "test_operation",
						Attributes: map[string]interface{}{
							"test.key": "test_value",
						},
					},
				},
			},
		}

		// Test validation
		inputData := generateTestTraces(t, contractDef.Inputs.Traces[0])
		outputData := generateTestTraces(t, contractDef.Inputs.Traces[0])

		validationResult := contractDef.Validate(inputData, outputData)
		assert.True(t, validationResult.Valid, "Contract validation should pass")
	})

	t.Run("MatcherFunctionality", func(t *testing.T) {
		// Test matcher functionality
		matcher := matcher.NewMatcher()

		contractDef := &contract.Contract{
			Publisher: "test-service",
			Pipeline:  "traces",
			Version:   "1.0",
			Inputs: contract.Inputs{
				Traces: []contract.TraceInput{
					{
						SpanName:    "test_operation",
						ServiceName: "test-service",
						Attributes: map[string]interface{}{
							"test.key": "test_value",
						},
					},
				},
			},
			Matchers: contract.Matchers{
				Traces: []contract.TraceMatcher{
					{
						SpanName: "test_operation",
						Attributes: map[string]interface{}{
							"test.key": "test_value",
						},
					},
				},
			},
		}

		inputData := generateTestTraces(t, contractDef.Inputs.Traces[0])
		outputData := generateTestTraces(t, contractDef.Inputs.Traces[0])

		validationResult := matcher.Validate(contractDef, inputData, outputData)
		assert.True(t, validationResult.Valid, "Matcher validation should pass")
	})
}

// TestIntegration_Performance tests performance characteristics
func TestIntegration_Performance(t *testing.T) {
	testCases := []struct {
		name         string
		numContracts int
		expectedTime time.Duration
		description  string
	}{
		{
			name:         "SmallScale",
			numContracts: 10,
			expectedTime: 5 * time.Second,
			description:  "Tests performance with 10 contracts",
		},
		{
			name:         "MediumScale",
			numContracts: 100,
			expectedTime: 30 * time.Second,
			description:  "Tests performance with 100 contracts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip large scale tests in CI to avoid timeouts
			if tc.numContracts > 50 && os.Getenv("CI") == "true" {
				t.Skip("Skipping large scale test in CI")
			}

			// Generate multiple contracts
			contracts := generateMultipleContracts(t, tc.numContracts)

			// Create test harness with minimal config
			config := harness.CollectorConfig{
				Receivers:  make(map[string]interface{}),
				Processors: make(map[string]interface{}),
				Exporters:  make(map[string]interface{}),
				Service:    make(map[string]interface{}),
			}

			testHarness := harness.NewTestHarness(harness.TestModePipeline, config)
			testHarness.SetLogger(createTestLogger())

			// Measure performance
			start := time.Now()
			results := testHarness.RunTests(contracts)
			duration := time.Since(start)

			assert.LessOrEqual(t, duration, tc.expectedTime,
				fmt.Sprintf("Test took %v, expected less than %v", duration, tc.expectedTime))

			// Validate results
			assert.Equal(t, tc.numContracts, results.TotalTests, "Expected all contracts to be tested")
			assert.Greater(t, results.PassedTests, 0, "Expected some tests to pass")
		})
	}
}

// TestIntegration_ErrorHandling tests error scenarios
func TestIntegration_ErrorHandling(t *testing.T) {
	t.Run("InvalidContract", func(t *testing.T) {
		// Test with invalid contract (missing required fields)
		contract := &contract.Contract{
			// Missing Publisher and Version
			Inputs: contract.Inputs{
				Traces: []contract.TraceInput{
					{
						SpanName: "test_operation",
					},
				},
			},
			Matchers: contract.Matchers{
				Traces: []contract.TraceMatcher{
					{
						SpanName: "test_operation",
					},
				},
			},
		}

		inputData := generateTestTraces(t, contract.Inputs.Traces[0])
		outputData := generateTestTraces(t, contract.Inputs.Traces[0])

		validationResult := contract.Validate(inputData, outputData)
		assert.False(t, validationResult.Valid, "Invalid contract should fail validation")
		assert.Greater(t, len(validationResult.Errors), 0, "Should have validation errors")
	})

	t.Run("EmptyOutputData", func(t *testing.T) {
		// Test with empty output data
		contractDef := &contract.Contract{
			Publisher: "test-service",
			Pipeline:  "traces",
			Version:   "1.0",
			Inputs: contract.Inputs{
				Traces: []contract.TraceInput{
					{
						SpanName: "test_operation",
					},
				},
			},
			Matchers: contract.Matchers{
				Traces: []contract.TraceMatcher{
					{
						SpanName: "test_operation",
					},
				},
			},
		}

		inputData := generateTestTraces(t, contractDef.Inputs.Traces[0])
		outputData := contract.OpenTelemetryData{
			Traces:  ptrace.NewTraces(),
			Metrics: pmetric.NewMetrics(),
			Logs:    plog.NewLogs(),
			Time:    time.Now(),
			// Empty traces
		}

		validationResult := contractDef.Validate(inputData, outputData)
		assert.False(t, validationResult.Valid, "Empty output should fail validation")
	})
}

// Helper functions

func createTestLogger() *zap.Logger {
	// Create a test logger that doesn't output to console
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}
	logger, _ := config.Build()
	return logger
}

func generateTestTraces(t *testing.T, input contract.TraceInput) contract.OpenTelemetryData {
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()

	// Set resource attributes
	if input.ServiceName != "" {
		resourceSpans.Resource().Attributes().PutStr("service.name", input.ServiceName)
	}

	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	span := scopeSpans.Spans().AppendEmpty()

	// Set span attributes
	span.SetName(input.SpanName)
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(time.Now().Add(100 * time.Millisecond)))

	// Set span attributes
	for key, value := range input.Attributes {
		switch v := value.(type) {
		case string:
			span.Attributes().PutStr(key, v)
		case int:
			span.Attributes().PutInt(key, int64(v))
		case float64:
			span.Attributes().PutDouble(key, v)
		case bool:
			span.Attributes().PutBool(key, v)
		default:
			span.Attributes().PutStr(key, fmt.Sprintf("%v", v))
		}
	}

	return contract.OpenTelemetryData{
		Traces: traces,
		Time:   time.Now(),
	}
}

func generateMultipleContracts(t *testing.T, count int) []*contract.Contract {
	contracts := make([]*contract.Contract, count)

	for i := 0; i < count; i++ {
		contracts[i] = &contract.Contract{
			Publisher: fmt.Sprintf("service-%d", i),
			Pipeline:  "traces",
			Version:   "1.0",
			Inputs: contract.Inputs{
				Traces: []contract.TraceInput{
					{
						SpanName:    fmt.Sprintf("operation-%d", i),
						ServiceName: fmt.Sprintf("service-%d", i),
						Attributes: map[string]interface{}{
							"test.key": fmt.Sprintf("value-%d", i),
						},
					},
				},
			},
			Matchers: contract.Matchers{
				Traces: []contract.TraceMatcher{
					{
						SpanName: fmt.Sprintf("operation-%d", i),
						Attributes: map[string]interface{}{
							"test.key": fmt.Sprintf("value-%d", i),
						},
					},
				},
			},
		}
	}

	return contracts
}
