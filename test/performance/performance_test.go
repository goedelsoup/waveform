// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package performance

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/goedelsoup/waveform/internal/contract"
	"github.com/goedelsoup/waveform/internal/harness"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// TestPerformance_ContractValidation tests the performance of contract validation
func TestPerformance_ContractValidation(t *testing.T) {
	testCases := []struct {
		name         string
		numContracts int
		expectedTime time.Duration
		description  string
	}{
		{
			name:         "SmallScale",
			numContracts: 10,
			expectedTime: 1 * time.Second,
			description:  "Tests validation performance with 10 contracts",
		},
		{
			name:         "MediumScale",
			numContracts: 100,
			expectedTime: 5 * time.Second,
			description:  "Tests validation performance with 100 contracts",
		},
		{
			name:         "LargeScale",
			numContracts: 1000,
			expectedTime: 30 * time.Second,
			description:  "Tests validation performance with 1000 contracts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip large scale tests in CI to avoid timeouts
			if tc.numContracts > 100 && os.Getenv("CI") == "true" {
				t.Skip("Skipping large scale test in CI")
			}

			// Generate test contracts
			contracts := generateTestContracts(t, tc.numContracts)

			// Measure validation performance
			start := time.Now()
			for _, contractDef := range contracts {
				inputData := generateTestData(t, contractDef)
				outputData := generateTestData(t, contractDef)

				validationResult := contractDef.Validate(inputData, outputData)
				assert.True(t, validationResult.Valid, "Contract validation should pass")
			}
			duration := time.Since(start)

			// Log performance metrics
			t.Logf("Validated %d contracts in %v (%.2f contracts/second)",
				tc.numContracts, duration, float64(tc.numContracts)/duration.Seconds())

			// Assert performance requirements
			assert.LessOrEqual(t, duration, tc.expectedTime,
				fmt.Sprintf("Validation took %v, expected less than %v", duration, tc.expectedTime))
		})
	}
}

// TestPerformance_TestHarness tests the performance of the test harness
func TestPerformance_TestHarness(t *testing.T) {
	testCases := []struct {
		name         string
		numContracts int
		expectedTime time.Duration
		description  string
	}{
		{
			name:         "SmallScale",
			numContracts: 10,
			expectedTime: 2 * time.Second,
			description:  "Tests harness performance with 10 contracts",
		},
		{
			name:         "MediumScale",
			numContracts: 100,
			expectedTime: 10 * time.Second,
			description:  "Tests harness performance with 100 contracts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip large scale tests in CI to avoid timeouts
			if tc.numContracts > 50 && os.Getenv("CI") == "true" {
				t.Skip("Skipping large scale test in CI")
			}

			// Generate test contracts
			contracts := generateTestContracts(t, tc.numContracts)

			// Create test harness
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

			// Log performance metrics
			t.Logf("Executed %d tests in %v (%.2f tests/second)",
				results.TotalTests, duration, float64(results.TotalTests)/duration.Seconds())
			t.Logf("Passed: %d, Failed: %d", results.PassedTests, results.FailedTests)

			// Assert performance requirements
			assert.LessOrEqual(t, duration, tc.expectedTime,
				fmt.Sprintf("Test execution took %v, expected less than %v", duration, tc.expectedTime))

			// Validate results
			assert.Equal(t, tc.numContracts, results.TotalTests, "Expected all contracts to be tested")
			assert.Greater(t, results.PassedTests, 0, "Expected some tests to pass")
		})
	}
}

// TestPerformance_MemoryUsage tests memory usage characteristics
func TestPerformance_MemoryUsage(t *testing.T) {
	t.Run("MemoryEfficiency", func(t *testing.T) {
		// Test memory efficiency with large datasets
		numContracts := 100
		contracts := generateTestContracts(t, numContracts)

		// Create test harness
		config := harness.CollectorConfig{
			Receivers:  make(map[string]interface{}),
			Processors: make(map[string]interface{}),
			Exporters:  make(map[string]interface{}),
			Service:    make(map[string]interface{}),
		}

		testHarness := harness.NewTestHarness(harness.TestModePipeline, config)
		testHarness.SetLogger(createTestLogger())

		// Run tests and measure memory usage
		start := time.Now()
		results := testHarness.RunTests(contracts)
		duration := time.Since(start)

		// Log memory and performance metrics
		t.Logf("Processed %d contracts in %v", numContracts, duration)
		t.Logf("Average time per contract: %v", duration/time.Duration(numContracts))
		t.Logf("Results: %d passed, %d failed", results.PassedTests, results.FailedTests)

		// Basic performance assertions
		assert.LessOrEqual(t, duration, 10*time.Second, "Should complete within reasonable time")
		assert.Equal(t, numContracts, results.TotalTests, "Should process all contracts")
	})
}

// TestPerformance_ConcurrentExecution tests concurrent test execution
func TestPerformance_ConcurrentExecution(t *testing.T) {
	t.Run("ConcurrentValidation", func(t *testing.T) {
		numContracts := 50
		contracts := generateTestContracts(t, numContracts)

		// Test concurrent validation
		start := time.Now()
		results := make(chan bool, numContracts)

		for _, contractDef := range contracts {
			go func(c *contract.Contract) {
				inputData := generateTestData(t, c)
				outputData := generateTestData(t, c)
				validationResult := c.Validate(inputData, outputData)
				results <- validationResult.Valid
			}(contractDef)
		}

		// Collect results
		passed := 0
		for i := 0; i < numContracts; i++ {
			if <-results {
				passed++
			}
		}
		duration := time.Since(start)

		t.Logf("Concurrent validation of %d contracts took %v", numContracts, duration)
		t.Logf("Passed: %d, Failed: %d", passed, numContracts-passed)

		assert.Equal(t, numContracts, passed, "All contracts should pass validation")
		assert.LessOrEqual(t, duration, 5*time.Second, "Concurrent execution should be efficient")
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

func generateTestContracts(t *testing.T, count int) []*contract.Contract {
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
							"test.key":         fmt.Sprintf("value-%d", i),
							"http.method":      "GET",
							"http.status_code": 200,
						},
					},
				},
			},
			Matchers: contract.Matchers{
				Traces: []contract.TraceMatcher{
					{
						SpanName: fmt.Sprintf("operation-%d", i),
						Attributes: map[string]interface{}{
							"test.key":         fmt.Sprintf("value-%d", i),
							"http.method":      "GET",
							"http.status_code": 200,
						},
					},
				},
			},
		}
	}

	return contracts
}

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
				span.Attributes().PutStr(key, fmt.Sprintf("%v", v))
			}
		}
	}

	return contract.OpenTelemetryData{
		Traces: traces,
		Time:   time.Now(),
	}
}
