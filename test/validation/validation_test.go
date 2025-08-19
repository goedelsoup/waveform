// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package validation

import (
	"testing"
	"time"

	"github.com/goedelsoup/waveform/internal/contract"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// TestValidation_ContractStructure tests contract structure validation
func TestValidation_ContractStructure(t *testing.T) {
	t.Run("ValidContract", func(t *testing.T) {
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

		inputData := generateTestData(t, contractDef)
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.True(t, validationResult.Valid, "Valid contract should pass validation")
		assert.Empty(t, validationResult.Errors, "Should have no validation errors")
	})

	t.Run("MissingPublisher", func(t *testing.T) {
		contractDef := &contract.Contract{
			// Missing Publisher
			Pipeline: "traces",
			Version:  "1.0",
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

		inputData := generateTestData(t, contractDef)
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.False(t, validationResult.Valid, "Contract missing publisher should fail validation")
		assert.Greater(t, len(validationResult.Errors), 0, "Should have validation errors")
		assert.Contains(t, validationResult.Errors[0].Message, "publisher is required")
	})

	t.Run("MissingVersion", func(t *testing.T) {
		contractDef := &contract.Contract{
			Publisher: "test-service",
			Pipeline:  "traces",
			// Missing Version
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

		inputData := generateTestData(t, contractDef)
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.False(t, validationResult.Valid, "Contract missing version should fail validation")
		assert.Greater(t, len(validationResult.Errors), 0, "Should have validation errors")
		assert.Contains(t, validationResult.Errors[0].Message, "version is required")
	})

	t.Run("MissingInputs", func(t *testing.T) {
		contractDef := &contract.Contract{
			Publisher: "test-service",
			Pipeline:  "traces",
			Version:   "1.0",
			// Missing Inputs
			Matchers: contract.Matchers{
				Traces: []contract.TraceMatcher{
					{
						SpanName: "test_operation",
					},
				},
			},
		}

		inputData := generateTestData(t, contractDef)
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.False(t, validationResult.Valid, "Contract missing inputs should fail validation")
		assert.Greater(t, len(validationResult.Errors), 0, "Should have validation errors")
		assert.Contains(t, validationResult.Errors[0].Message, "at least one input")
	})

	t.Run("MissingMatchers", func(t *testing.T) {
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
			// Missing Matchers
		}

		inputData := generateTestData(t, contractDef)
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.False(t, validationResult.Valid, "Contract missing matchers should fail validation")
		assert.Greater(t, len(validationResult.Errors), 0, "Should have validation errors")
		assert.Contains(t, validationResult.Errors[0].Message, "at least one matcher")
	})
}

// TestValidation_DataPresence tests data presence validation
func TestValidation_DataPresence(t *testing.T) {
	t.Run("ValidInputData", func(t *testing.T) {
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

		inputData := generateTestData(t, contractDef)
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.True(t, validationResult.Valid, "Valid input data should pass validation")
	})

	t.Run("EmptyInputData", func(t *testing.T) {
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

		// Empty input data
		inputData := contract.OpenTelemetryData{
			Traces:  ptrace.NewTraces(),
			Metrics: pmetric.NewMetrics(),
			Logs:    plog.NewLogs(),
			Time:    time.Now(),
		}
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.False(t, validationResult.Valid, "Empty input data should fail validation")
		assert.Greater(t, len(validationResult.Errors), 0, "Should have validation errors")
		assert.Contains(t, validationResult.Errors[0].Message, "no traces provided")
	})

	t.Run("EmptyOutputData", func(t *testing.T) {
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

		inputData := generateTestData(t, contractDef)
		// Empty output data
		outputData := contract.OpenTelemetryData{
			Traces:  ptrace.NewTraces(),
			Metrics: pmetric.NewMetrics(),
			Logs:    plog.NewLogs(),
			Time:    time.Now(),
		}

		validationResult := contractDef.Validate(inputData, outputData)
		assert.False(t, validationResult.Valid, "Empty output data should fail validation")
		assert.Greater(t, len(validationResult.Errors), 0, "Should have validation errors")
		assert.Contains(t, validationResult.Errors[0].Message, "no traces found")
	})
}

// TestValidation_TimeWindows tests time window validation
func TestValidation_TimeWindows(t *testing.T) {
	t.Run("NoTimeWindows", func(t *testing.T) {
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
			// No time windows
		}

		inputData := generateTestData(t, contractDef)
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.True(t, validationResult.Valid, "Contract without time windows should pass validation")
	})

	t.Run("WithTimeWindows", func(t *testing.T) {
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
			TimeWindows: []contract.TimeWindow{
				{
					Duration: "1h",
				},
			},
		}

		inputData := generateTestData(t, contractDef)
		outputData := generateTestData(t, contractDef)

		validationResult := contractDef.Validate(inputData, outputData)
		assert.True(t, validationResult.Valid, "Contract with time windows should pass validation")
	})
}

// Helper functions

func generateTestData(t *testing.T, contractDef *contract.Contract) contract.OpenTelemetryData {
	// Generate test data based on contract inputs
	traces := ptrace.NewTraces()

	if len(contractDef.Inputs.Traces) > 0 {
		input := contractDef.Inputs.Traces[0]
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
				span.Attributes().PutStr(key, v.(string))
			}
		}
	}

	return contract.OpenTelemetryData{
		Traces: traces,
		Time:   time.Now(),
	}
}
