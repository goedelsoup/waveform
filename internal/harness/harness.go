// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package harness

import (
	"context"
	"fmt"
	"time"

	"github.com/goedelsoup/waveform/internal/contract"
	"github.com/goedelsoup/waveform/internal/generator"
	"github.com/goedelsoup/waveform/internal/matcher"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// TestMode represents the testing mode
type TestMode string

const (
	TestModePipeline  TestMode = "pipeline"
	TestModeProcessor TestMode = "processor"
)

// TestResult represents the result of a single test
type TestResult struct {
	Contract   *contract.Contract
	Valid      bool
	Errors     []string
	Warnings   []string
	Duration   time.Duration
	InputData  contract.OpenTelemetryData
	OutputData contract.OpenTelemetryData
}

// TestResults represents the results of all tests
type TestResults struct {
	Results     []TestResult
	TotalTests  int
	PassedTests int
	FailedTests int
	Duration    time.Duration
}

// CollectorConfig represents the configuration for a collector
type CollectorConfig struct {
	Receivers  map[string]interface{} `yaml:"receivers"`
	Processors map[string]interface{} `yaml:"processors"`
	Exporters  map[string]interface{} `yaml:"exporters"`
	Service    map[string]interface{} `yaml:"service"`
}

// TestHarness orchestrates test execution
type TestHarness struct {
	mode      TestMode
	config    CollectorConfig
	generator *generator.Generator
	matcher   *matcher.Matcher
	logger    *zap.Logger
}

// NewTestHarness creates a new test harness
func NewTestHarness(mode TestMode, config CollectorConfig) *TestHarness {
	return &TestHarness{
		mode:      mode,
		config:    config,
		generator: generator.NewGenerator(),
		matcher:   matcher.NewMatcher(),
		logger:    zap.NewNop(),
	}
}

// SetLogger sets the logger for the test harness
func (h *TestHarness) SetLogger(logger *zap.Logger) {
	h.logger = logger
}

// RunTests runs all tests for the given contracts
func (h *TestHarness) RunTests(contracts []*contract.Contract) TestResults {
	startTime := time.Now()
	results := TestResults{
		Results: make([]TestResult, 0, len(contracts)),
	}

	h.logger.Info("Starting test execution",
		zap.String("mode", string(h.mode)),
		zap.Int("contract_count", len(contracts)))

	for _, contract := range contracts {
		result := h.runSingleTest(contract)
		results.Results = append(results.Results, result)

		if result.Valid {
			results.PassedTests++
		} else {
			results.FailedTests++
		}
	}

	results.TotalTests = len(contracts)
	results.Duration = time.Since(startTime)

	h.logger.Info("Test execution completed",
		zap.Int("total", results.TotalTests),
		zap.Int("passed", results.PassedTests),
		zap.Int("failed", results.FailedTests),
		zap.Duration("duration", results.Duration))

	return results
}

// runSingleTest runs a single test for a contract
func (h *TestHarness) runSingleTest(contractDef *contract.Contract) TestResult {
	startTime := time.Now()
	result := TestResult{
		Contract: contractDef,
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	h.logger.Debug("Running test",
		zap.String("publisher", contractDef.Publisher),
		zap.String("pipeline", contractDef.Pipeline))

	// Generate input data
	inputData := h.generator.GenerateFromContract(contractDef)
	result.InputData = inputData

	// Run the test based on mode
	var outputData contract.OpenTelemetryData
	var err error

	switch h.mode {
	case TestModePipeline:
		outputData, err = h.runPipelineTest(contractDef, inputData)
	case TestModeProcessor:
		outputData, err = h.runProcessorTest(contractDef, inputData)
	default:
		err = fmt.Errorf("unknown test mode: %s", h.mode)
	}

	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		result.Duration = time.Since(startTime)
		return result
	}

	result.OutputData = outputData

	// Validate the output
	validationResult := h.matcher.Validate(contractDef, inputData, outputData)
	result.Valid = validationResult.Valid

	for _, validationError := range validationResult.Errors {
		result.Errors = append(result.Errors, validationError.Message)
	}

	for _, warning := range validationResult.Warnings {
		result.Warnings = append(result.Warnings, warning)
	}

	result.Duration = time.Since(startTime)

	if result.Valid {
		h.logger.Debug("Test passed",
			zap.String("publisher", contractDef.Publisher),
			zap.String("pipeline", contractDef.Pipeline),
			zap.Duration("duration", result.Duration))
	} else {
		h.logger.Warn("Test failed",
			zap.String("publisher", contractDef.Publisher),
			zap.String("pipeline", contractDef.Pipeline),
			zap.Strings("errors", result.Errors),
			zap.Duration("duration", result.Duration))
	}

	return result
}

// runPipelineTest runs a test in pipeline mode
func (h *TestHarness) runPipelineTest(contractDef *contract.Contract, inputData contract.OpenTelemetryData) (contract.OpenTelemetryData, error) {
	// In a real implementation, this would:
	// 1. Start a collector instance with the provided configuration
	// 2. Send input data through the pipeline
	// 3. Capture output data from exporters
	// 4. Return the output data

	// For now, we'll simulate the pipeline by returning the input data
	// This is a placeholder implementation
	h.logger.Debug("Running pipeline test",
		zap.String("publisher", contractDef.Publisher),
		zap.String("pipeline", contractDef.Pipeline))

	// Simulate some processing delay
	time.Sleep(10 * time.Millisecond)

	// Return the input data as output (no transformation for now)
	return inputData, nil
}

// runProcessorTest runs a test in processor mode
func (h *TestHarness) runProcessorTest(contractDef *contract.Contract, inputData contract.OpenTelemetryData) (contract.OpenTelemetryData, error) {
	// In a real implementation, this would:
	// 1. Create processor instances based on the configuration
	// 2. Process input data through each processor
	// 3. Return the processed data

	// For now, we'll simulate processor testing by returning the input data
	// This is a placeholder implementation
	h.logger.Debug("Running processor test",
		zap.String("publisher", contractDef.Publisher),
		zap.String("pipeline", contractDef.Pipeline))

	// Simulate some processing delay
	time.Sleep(5 * time.Millisecond)

	// Return the input data as output (no transformation for now)
	return inputData, nil
}

// MockConsumer is a mock consumer for testing
type MockConsumer struct {
	traces  []ptrace.Traces
	metrics []pmetric.Metrics
	logs    []plog.Logs
}

// NewMockConsumer creates a new mock consumer
func NewMockConsumer() *MockConsumer {
	return &MockConsumer{
		traces:  make([]ptrace.Traces, 0),
		metrics: make([]pmetric.Metrics, 0),
		logs:    make([]plog.Logs, 0),
	}
}

// Capabilities returns the consumer capabilities
func (m *MockConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// ConsumeTraces consumes trace data
func (m *MockConsumer) ConsumeTraces(ctx context.Context, traces ptrace.Traces) error {
	m.traces = append(m.traces, traces)
	return nil
}

// ConsumeMetrics consumes metric data
func (m *MockConsumer) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	m.metrics = append(m.metrics, metrics)
	return nil
}

// ConsumeLogs consumes log data
func (m *MockConsumer) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	m.logs = append(m.logs, logs)
	return nil
}

// GetTraces returns all consumed traces
func (m *MockConsumer) GetTraces() []ptrace.Traces {
	return m.traces
}

// GetMetrics returns all consumed metrics
func (m *MockConsumer) GetMetrics() []pmetric.Metrics {
	return m.metrics
}

// GetLogs returns all consumed logs
func (m *MockConsumer) GetLogs() []plog.Logs {
	return m.logs
}

// Clear clears all consumed data
func (m *MockConsumer) Clear() {
	m.traces = make([]ptrace.Traces, 0)
	m.metrics = make([]pmetric.Metrics, 0)
	m.logs = make([]plog.Logs, 0)
}

// ProcessorFactory creates processor instances for testing
type ProcessorFactory struct {
	logger *zap.Logger
}

// NewProcessorFactory creates a new processor factory
func NewProcessorFactory(logger *zap.Logger) *ProcessorFactory {
	return &ProcessorFactory{
		logger: logger,
	}
}

// CreateProcessor creates a processor instance
func (f *ProcessorFactory) CreateProcessor(processorType string, config component.Config) (component.Component, error) {
	// This is a placeholder implementation
	// In a real implementation, this would create actual processor instances
	f.logger.Debug("Creating processor", zap.String("type", processorType))

	// Return a mock processor for now
	return &MockProcessor{}, nil
}

// MockProcessor is a mock processor for testing
type MockProcessor struct {
	component.StartFunc
	component.ShutdownFunc
}

// Capabilities returns the processor capabilities
func (m *MockProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// ConsumeTraces processes trace data
func (m *MockProcessor) ConsumeTraces(ctx context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	// For now, just return the input traces unchanged
	return traces, nil
}

// ConsumeMetrics processes metric data
func (m *MockProcessor) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	// For now, just return the input metrics unchanged
	return metrics, nil
}

// ConsumeLogs processes log data
func (m *MockProcessor) ConsumeLogs(ctx context.Context, logs plog.Logs) (plog.Logs, error) {
	// For now, just return the input logs unchanged
	return logs, nil
}
