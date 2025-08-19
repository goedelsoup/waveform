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
	mode             TestMode
	config           CollectorConfig
	generator        *generator.Generator
	matcher          *matcher.Matcher
	logger           *zap.Logger
	collectorService CollectorService
}

// NewTestHarness creates a new test harness
func NewTestHarness(mode TestMode, config CollectorConfig) *TestHarness {
	return &TestHarness{
		mode:             mode,
		config:           config,
		generator:        generator.NewGenerator(),
		matcher:          matcher.NewMatcher(),
		logger:           zap.NewNop(),
		collectorService: nil, // Will be set when needed
	}
}

// SetLogger sets the logger for the test harness
func (h *TestHarness) SetLogger(logger *zap.Logger) {
	h.logger = logger
}

// SetCollectorService sets the collector service for the test harness
func (h *TestHarness) SetCollectorService(service CollectorService) {
	h.collectorService = service
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
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	h.logger.Debug("Running test",
		zap.String("publisher", contractDef.Publisher),
		zap.String("pipeline", contractDef.Pipeline),
		zap.String("version", contractDef.Version))

	// Generate input data from contract
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
		result.Errors = append(result.Errors, fmt.Sprintf("Unknown test mode: %s", h.mode))
		result.Valid = false
		result.Duration = time.Since(startTime)
		return result
	}

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Test execution failed: %v", err))
		result.Valid = false
		result.Duration = time.Since(startTime)
		return result
	}

	result.OutputData = outputData

	// Validate the output against contract matchers
	validationResult := h.matcher.Validate(contractDef, inputData, outputData)
	if !validationResult.Valid {
		for _, err := range validationResult.Errors {
			result.Errors = append(result.Errors, err.Message)
		}
		result.Valid = false
	} else {
		result.Valid = true
	}

	result.Duration = time.Since(startTime)

	h.logger.Debug("Test completed",
		zap.String("publisher", contractDef.Publisher),
		zap.Bool("valid", result.Valid),
		zap.Int("errors", len(result.Errors)),
		zap.Duration("duration", result.Duration))

	return result
}

// runPipelineTest runs a test in pipeline mode
func (h *TestHarness) runPipelineTest(contractDef *contract.Contract, inputData contract.OpenTelemetryData) (contract.OpenTelemetryData, error) {
	h.logger.Debug("Running pipeline test",
		zap.String("publisher", contractDef.Publisher),
		zap.String("pipeline", contractDef.Pipeline))

	// Use collector service if available, otherwise fall back to simulation
	if h.collectorService != nil {
		return h.runPipelineTestWithCollector(contractDef, inputData)
	}

	// Fall back to simulation mode
	return h.runPipelineTestSimulation(contractDef, inputData)
}

// runPipelineTestWithCollector runs a test using the real collector service
func (h *TestHarness) runPipelineTestWithCollector(contractDef *contract.Contract, inputData contract.OpenTelemetryData) (contract.OpenTelemetryData, error) {
	h.logger.Debug("Running pipeline test with real collector service")

	// Start the collector service
	if err := h.collectorService.Start(); err != nil {
		return contract.OpenTelemetryData{}, fmt.Errorf("failed to start collector service: %w", err)
	}
	defer func() {
		if err := h.collectorService.Stop(); err != nil {
			h.logger.Error("Failed to stop collector service", zap.Error(err))
		}
	}()

	// Process data through the collector
	output, err := h.collectorService.ProcessData(inputData)
	if err != nil {
		return contract.OpenTelemetryData{}, fmt.Errorf("failed to process data through collector: %w", err)
	}

	// Convert output back to OpenTelemetryData
	if outputData, ok := output.(contract.OpenTelemetryData); ok {
		return outputData, nil
	}

	// If conversion fails, return empty data
	return contract.OpenTelemetryData{
		Time:    time.Now(),
		Traces:  ptrace.NewTraces(),
		Metrics: pmetric.NewMetrics(),
		Logs:    plog.NewLogs(),
	}, nil
}

// runPipelineTestSimulation runs a test using simulation mode
func (h *TestHarness) runPipelineTestSimulation(contractDef *contract.Contract, inputData contract.OpenTelemetryData) (contract.OpenTelemetryData, error) {
	h.logger.Debug("Running pipeline test in simulation mode")

	// Create a mock consumer to capture output data
	mockConsumer := NewMockConsumer()
	defer mockConsumer.Clear()

	// Create a simple pipeline simulation
	ctx := context.Background()

	// Process traces if present
	if inputData.Traces.ResourceSpans().Len() > 0 {
		h.logger.Debug("Processing traces through pipeline")
		if err := h.processTracesThroughPipeline(ctx, inputData.Traces, mockConsumer); err != nil {
			return contract.OpenTelemetryData{}, fmt.Errorf("failed to process traces: %w", err)
		}
	}

	// Process metrics if present (only if metrics exist and have data)
	if inputData.Metrics.ResourceMetrics().Len() > 0 {
		h.logger.Debug("Processing metrics through pipeline")
		if err := h.processMetricsThroughPipeline(ctx, inputData.Metrics, mockConsumer); err != nil {
			return contract.OpenTelemetryData{}, fmt.Errorf("failed to process metrics: %w", err)
		}
	} else {
		h.logger.Debug("No metrics to process")
	}

	// Process logs if present (only if logs exist and have data)
	if inputData.Logs.ResourceLogs().Len() > 0 {
		h.logger.Debug("Processing logs through pipeline")
		if err := h.processLogsThroughPipeline(ctx, inputData.Logs, mockConsumer); err != nil {
			return contract.OpenTelemetryData{}, fmt.Errorf("failed to process logs: %w", err)
		}
	} else {
		h.logger.Debug("No logs to process")
	}

	// Collect output data
	outputData := contract.OpenTelemetryData{
		Time:    time.Now(),
		Traces:  ptrace.NewTraces(),
		Metrics: pmetric.NewMetrics(),
		Logs:    plog.NewLogs(),
	}

	// Get processed traces
	if traces := mockConsumer.GetTraces(); len(traces) > 0 {
		outputData.Traces = traces[0] // Take the first batch for now
	}

	// Get processed metrics
	if metrics := mockConsumer.GetMetrics(); len(metrics) > 0 {
		outputData.Metrics = metrics[0] // Take the first batch for now
	}

	// Get processed logs
	if logs := mockConsumer.GetLogs(); len(logs) > 0 {
		outputData.Logs = logs[0] // Take the first batch for now
	}

	return outputData, nil
}

// runProcessorTest runs a test in processor mode
func (h *TestHarness) runProcessorTest(contractDef *contract.Contract, inputData contract.OpenTelemetryData) (contract.OpenTelemetryData, error) {
	h.logger.Debug("Running processor test",
		zap.String("publisher", contractDef.Publisher),
		zap.String("pipeline", contractDef.Pipeline))

	// Create a mock consumer to capture output data
	mockConsumer := NewMockConsumer()
	defer mockConsumer.Clear()

	ctx := context.Background()

	// Process traces if present
	if inputData.Traces.ResourceSpans().Len() > 0 {
		h.logger.Debug("Processing traces through processors")
		if err := h.processTracesThroughProcessors(ctx, inputData.Traces, mockConsumer); err != nil {
			return contract.OpenTelemetryData{}, fmt.Errorf("failed to process traces: %w", err)
		}
	}

	// Process metrics if present (only if metrics exist and have data)
	if inputData.Metrics.ResourceMetrics().Len() > 0 {
		h.logger.Debug("Processing metrics through processors")
		if err := h.processMetricsThroughProcessors(ctx, inputData.Metrics, mockConsumer); err != nil {
			return contract.OpenTelemetryData{}, fmt.Errorf("failed to process metrics: %w", err)
		}
	} else {
		h.logger.Debug("No metrics to process")
	}

	// Process logs if present (only if logs exist and have data)
	if inputData.Logs.ResourceLogs().Len() > 0 {
		h.logger.Debug("Processing logs through processors")
		if err := h.processLogsThroughProcessors(ctx, inputData.Logs, mockConsumer); err != nil {
			return contract.OpenTelemetryData{}, fmt.Errorf("failed to process logs: %w", err)
		}
	} else {
		h.logger.Debug("No logs to process")
	}

	// Collect output data
	outputData := contract.OpenTelemetryData{
		Time:    time.Now(),
		Traces:  ptrace.NewTraces(),
		Metrics: pmetric.NewMetrics(),
		Logs:    plog.NewLogs(),
	}

	// Get processed traces
	if traces := mockConsumer.GetTraces(); len(traces) > 0 {
		outputData.Traces = traces[0]
	}

	// Get processed metrics
	if metrics := mockConsumer.GetMetrics(); len(metrics) > 0 {
		outputData.Metrics = metrics[0]
	}

	// Get processed logs
	if logs := mockConsumer.GetLogs(); len(logs) > 0 {
		outputData.Logs = logs[0]
	}

	return outputData, nil
}

// processTracesThroughPipeline processes traces through the configured pipeline
func (h *TestHarness) processTracesThroughPipeline(ctx context.Context, traces ptrace.Traces, consumer *MockConsumer) error {
	// In a real implementation, this would:
	// 1. Start a collector instance with the provided configuration
	// 2. Send traces through the configured pipeline
	// 3. Capture output from exporters

	// For now, we'll simulate basic processing based on the configuration
	h.logger.Debug("Simulating trace processing through pipeline")

	// Apply basic transformations based on processor configuration
	processedTraces := h.applyTraceTransformations(traces)

	// Send to consumer
	return consumer.ConsumeTraces(ctx, processedTraces)
}

// processMetricsThroughPipeline processes metrics through the configured pipeline
func (h *TestHarness) processMetricsThroughPipeline(ctx context.Context, metrics pmetric.Metrics, consumer *MockConsumer) error {
	h.logger.Debug("Simulating metric processing through pipeline")

	// Apply basic transformations based on processor configuration
	processedMetrics := h.applyMetricTransformations(metrics)

	// Send to consumer
	return consumer.ConsumeMetrics(ctx, processedMetrics)
}

// processLogsThroughPipeline processes logs through the configured pipeline
func (h *TestHarness) processLogsThroughPipeline(ctx context.Context, logs plog.Logs, consumer *MockConsumer) error {
	h.logger.Debug("Simulating log processing through pipeline")

	// Apply basic transformations based on processor configuration
	processedLogs := h.applyLogTransformations(logs)

	// Send to consumer
	return consumer.ConsumeLogs(ctx, processedLogs)
}

// processTracesThroughProcessors processes traces through individual processors
func (h *TestHarness) processTracesThroughProcessors(ctx context.Context, traces ptrace.Traces, consumer *MockConsumer) error {
	h.logger.Debug("Simulating trace processing through processors")

	// Apply processor-specific transformations
	processedTraces := h.applyTraceTransformations(traces)

	// Send to consumer
	return consumer.ConsumeTraces(ctx, processedTraces)
}

// processMetricsThroughProcessors processes metrics through individual processors
func (h *TestHarness) processMetricsThroughProcessors(ctx context.Context, metrics pmetric.Metrics, consumer *MockConsumer) error {
	h.logger.Debug("Simulating metric processing through processors")

	// Apply processor-specific transformations
	processedMetrics := h.applyMetricTransformations(metrics)

	// Send to consumer
	return consumer.ConsumeMetrics(ctx, processedMetrics)
}

// processLogsThroughProcessors processes logs through individual processors
func (h *TestHarness) processLogsThroughProcessors(ctx context.Context, logs plog.Logs, consumer *MockConsumer) error {
	h.logger.Debug("Simulating log processing through processors")

	// Apply processor-specific transformations
	processedLogs := h.applyLogTransformations(logs)

	// Send to consumer
	return consumer.ConsumeLogs(ctx, processedLogs)
}

// applyTraceTransformations applies transformations to traces based on processor configuration
func (h *TestHarness) applyTraceTransformations(traces ptrace.Traces) ptrace.Traces {
	// Create a copy to avoid modifying the original
	processedTraces := ptrace.NewTraces()
	traces.CopyTo(processedTraces)

	// Apply transformations based on processor configuration
	// This is a simplified implementation - in a real scenario, you would:
	// 1. Parse the processor configuration
	// 2. Create processor instances
	// 3. Apply each processor in sequence

	h.logger.Debug("Applying trace transformations", zap.Int("processor_count", len(h.config.Processors)))

	// For now, we'll apply some basic transformations based on common processors
	for processorName, processorConfig := range h.config.Processors {
		h.logger.Debug("Processing with processor", zap.String("processor", processorName))

		// Apply transformations based on processor type
		switch processorName {
		case "transform":
			processedTraces = h.applyTransformProcessor(processedTraces, processorConfig)
		case "attributes":
			processedTraces = h.applyAttributesProcessor(processedTraces, processorConfig)
		case "filter":
			processedTraces = h.applyFilterProcessor(processedTraces, processorConfig)
		default:
			h.logger.Debug("Unknown processor type, skipping", zap.String("processor", processorName))
		}
	}

	return processedTraces
}

// applyMetricTransformations applies transformations to metrics based on processor configuration
func (h *TestHarness) applyMetricTransformations(metrics pmetric.Metrics) pmetric.Metrics {
	// Create a copy to avoid modifying the original
	processedMetrics := pmetric.NewMetrics()
	metrics.CopyTo(processedMetrics)

	h.logger.Debug("Applying metric transformations", zap.Int("processor_count", len(h.config.Processors)))

	// Apply transformations based on processor configuration
	for processorName, processorConfig := range h.config.Processors {
		h.logger.Debug("Processing with processor", zap.String("processor", processorName))

		switch processorName {
		case "transform":
			processedMetrics = h.applyMetricTransformProcessor(processedMetrics, processorConfig)
		case "attributes":
			processedMetrics = h.applyMetricAttributesProcessor(processedMetrics, processorConfig)
		case "filter":
			processedMetrics = h.applyMetricFilterProcessor(processedMetrics, processorConfig)
		default:
			h.logger.Debug("Unknown processor type, skipping", zap.String("processor", processorName))
		}
	}

	return processedMetrics
}

// applyLogTransformations applies transformations to logs based on processor configuration
func (h *TestHarness) applyLogTransformations(logs plog.Logs) plog.Logs {
	// Create a copy to avoid modifying the original
	processedLogs := plog.NewLogs()
	logs.CopyTo(processedLogs)

	h.logger.Debug("Applying log transformations", zap.Int("processor_count", len(h.config.Processors)))

	// Apply transformations based on processor configuration
	for processorName, processorConfig := range h.config.Processors {
		h.logger.Debug("Processing with processor", zap.String("processor", processorName))

		switch processorName {
		case "transform":
			processedLogs = h.applyLogTransformProcessor(processedLogs, processorConfig)
		case "attributes":
			processedLogs = h.applyLogAttributesProcessor(processedLogs, processorConfig)
		case "filter":
			processedLogs = h.applyLogFilterProcessor(processedLogs, processorConfig)
		default:
			h.logger.Debug("Unknown processor type, skipping", zap.String("processor", processorName))
		}
	}

	return processedLogs
}

// Processor transformation methods (simplified implementations)
func (h *TestHarness) applyTransformProcessor(traces ptrace.Traces, config interface{}) ptrace.Traces {
	// Simplified transform processor implementation
	// In a real implementation, this would parse the config and apply transformations
	h.logger.Debug("Applying transform processor to traces")
	return traces
}

func (h *TestHarness) applyAttributesProcessor(traces ptrace.Traces, config interface{}) ptrace.Traces {
	// Simplified attributes processor implementation
	h.logger.Debug("Applying attributes processor to traces")
	return traces
}

func (h *TestHarness) applyFilterProcessor(traces ptrace.Traces, config interface{}) ptrace.Traces {
	// Simplified filter processor implementation
	h.logger.Debug("Applying filter processor to traces")
	return traces
}

func (h *TestHarness) applyMetricTransformProcessor(metrics pmetric.Metrics, config interface{}) pmetric.Metrics {
	h.logger.Debug("Applying transform processor to metrics")
	return metrics
}

func (h *TestHarness) applyMetricAttributesProcessor(metrics pmetric.Metrics, config interface{}) pmetric.Metrics {
	h.logger.Debug("Applying attributes processor to metrics")
	return metrics
}

func (h *TestHarness) applyMetricFilterProcessor(metrics pmetric.Metrics, config interface{}) pmetric.Metrics {
	h.logger.Debug("Applying filter processor to metrics")
	return metrics
}

func (h *TestHarness) applyLogTransformProcessor(logs plog.Logs, config interface{}) plog.Logs {
	h.logger.Debug("Applying transform processor to logs")
	return logs
}

func (h *TestHarness) applyLogAttributesProcessor(logs plog.Logs, config interface{}) plog.Logs {
	h.logger.Debug("Applying attributes processor to logs")
	return logs
}

func (h *TestHarness) applyLogFilterProcessor(logs plog.Logs, config interface{}) plog.Logs {
	h.logger.Debug("Applying filter processor to logs")
	return logs
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
