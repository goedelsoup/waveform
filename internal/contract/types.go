// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package contract

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// SignalType represents the type of telemetry signal
type SignalType string

const (
	SignalTypeTraces  SignalType = "traces"
	SignalTypeMetrics SignalType = "metrics"
	SignalTypeLogs    SignalType = "logs"
)

// FilterOperator represents the type of filter operation
type FilterOperator string

const (
	FilterOperatorEquals      FilterOperator = "equals"
	FilterOperatorNotEquals   FilterOperator = "not_equals"
	FilterOperatorMatches     FilterOperator = "matches"
	FilterOperatorExists      FilterOperator = "exists"
	FilterOperatorNotExists   FilterOperator = "not_exists"
	FilterOperatorGreaterThan FilterOperator = "greater_than"
	FilterOperatorLessThan    FilterOperator = "less_than"
)

// PipelineSelectorOperator represents the type of pipeline selector operation
type PipelineSelectorOperator string

const (
	PipelineSelectorOperatorEquals     PipelineSelectorOperator = "equals"
	PipelineSelectorOperatorMatches    PipelineSelectorOperator = "matches"
	PipelineSelectorOperatorContains   PipelineSelectorOperator = "contains"
	PipelineSelectorOperatorStartsWith PipelineSelectorOperator = "starts_with"
	PipelineSelectorOperatorEndsWith   PipelineSelectorOperator = "ends_with"
)

// PipelineSelector represents a criterion for matching pipelines
type PipelineSelector struct {
	Field    string                   `yaml:"field"`
	Operator PipelineSelectorOperator `yaml:"operator"`
	Value    interface{}              `yaml:"value"`
}

// PipelineSelectors represents a set of criteria for matching pipelines
type PipelineSelectors struct {
	Selectors []PipelineSelector `yaml:"selectors,omitempty"`
	Priority  int                `yaml:"priority,omitempty"` // Higher priority selectors are preferred
}

// Filter represents a filter predicate for determining when to apply a contract
type Filter struct {
	Field    string         `yaml:"field"`
	Operator FilterOperator `yaml:"operator"`
	Value    interface{}    `yaml:"value"`
}

// TimeWindow represents timing-sensitive transformations
type TimeWindow struct {
	Aggregation      string `yaml:"aggregation"`
	Duration         string `yaml:"duration"`
	ExpectedBehavior string `yaml:"expected_behavior"`
}

// TraceInput represents input trace data
type TraceInput struct {
	SpanName    string                 `yaml:"span_name"`
	Attributes  map[string]interface{} `yaml:"attributes,omitempty"`
	ParentSpan  string                 `yaml:"parent_span,omitempty"`
	ServiceName string                 `yaml:"service_name,omitempty"`
}

// MetricInput represents input metric data
type MetricInput struct {
	Name   string                 `yaml:"name"`
	Value  interface{}            `yaml:"value"`
	Type   string                 `yaml:"type,omitempty"` // counter, gauge, histogram
	Labels map[string]interface{} `yaml:"labels,omitempty"`
}

// LogInput represents input log data
type LogInput struct {
	Body       string                 `yaml:"body"`
	Severity   string                 `yaml:"severity,omitempty"`
	Attributes map[string]interface{} `yaml:"attributes,omitempty"`
}

// TraceMatcher represents expected trace transformations
type TraceMatcher struct {
	SpanName    string                 `yaml:"span_name,omitempty"`
	Attributes  map[string]interface{} `yaml:"attributes,omitempty"`
	ParentSpan  string                 `yaml:"parent_span,omitempty"`
	ServiceName string                 `yaml:"service_name,omitempty"`
}

// MetricMatcher represents expected metric transformations
type MetricMatcher struct {
	Name   string                 `yaml:"name,omitempty"`
	Type   string                 `yaml:"type,omitempty"`
	Labels map[string]interface{} `yaml:"labels,omitempty"`
}

// LogMatcher represents expected log transformations
type LogMatcher struct {
	Body       string                 `yaml:"body,omitempty"`
	Severity   string                 `yaml:"severity,omitempty"`
	Attributes map[string]interface{} `yaml:"attributes,omitempty"`
}

// Inputs represents the input data samples or generation rules
type Inputs struct {
	Traces  []TraceInput  `yaml:"traces,omitempty"`
	Metrics []MetricInput `yaml:"metrics,omitempty"`
	Logs    []LogInput    `yaml:"logs,omitempty"`
}

// Matchers represents expected transformation matchers
type Matchers struct {
	Traces  []TraceMatcher  `yaml:"traces,omitempty"`
	Metrics []MetricMatcher `yaml:"metrics,omitempty"`
	Logs    []LogMatcher    `yaml:"logs,omitempty"`
}

// Contract represents a complete contract definition
type Contract struct {
	Publisher         string             `yaml:"publisher"`
	Pipeline          string             `yaml:"pipeline,omitempty"`           // Explicit pipeline ID (deprecated in favor of selectors)
	PipelineSelectors *PipelineSelectors `yaml:"pipeline_selectors,omitempty"` // Pipeline matching criteria
	Version           string             `yaml:"version"`
	Description       string             `yaml:"description,omitempty"`
	Inputs            Inputs             `yaml:"inputs"`
	Filters           []Filter           `yaml:"filters,omitempty"`
	Matchers          Matchers           `yaml:"matchers"`
	TimeWindows       []TimeWindow       `yaml:"time_windows,omitempty"`
	FilePath          string             `yaml:"-"` // Set by loader
}

// OpenTelemetryData represents the unified data structure for all signal types
type OpenTelemetryData struct {
	Traces  ptrace.Traces
	Metrics pmetric.Metrics
	Logs    plog.Logs
	Time    time.Time
}

// ValidationResult represents the result of contract validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []string
}

// ValidationError represents a specific validation error
type ValidationError struct {
	Type       string
	Message    string
	Field      string
	Expected   interface{}
	Actual     interface{}
	SignalType SignalType
	Index      int
}

// Contract interface for validation
type ContractValidator interface {
	Validate(input, output OpenTelemetryData) ValidationResult
	GetPublisher() string
	GetPipeline() string
	GetVersion() string
	GetFilters() []Filter
	GetMatchers() Matchers
}

// Ensure Contract implements ContractValidator
func (c *Contract) Validate(input, output OpenTelemetryData) ValidationResult {
	// This will be implemented in the validation engine
	return ValidationResult{Valid: true}
}

func (c *Contract) GetPublisher() string {
	return c.Publisher
}

// GetPipeline returns the explicit pipeline ID or empty string if using selectors
func (c *Contract) GetPipeline() string {
	return c.Pipeline
}

// GetPipelineSelectors returns the pipeline selectors if defined
func (c *Contract) GetPipelineSelectors() *PipelineSelectors {
	return c.PipelineSelectors
}

// HasPipelineSelectors returns true if the contract uses pipeline selectors
func (c *Contract) HasPipelineSelectors() bool {
	return c.PipelineSelectors != nil && len(c.PipelineSelectors.Selectors) > 0
}

func (c *Contract) GetVersion() string {
	return c.Version
}

func (c *Contract) GetFilters() []Filter {
	return c.Filters
}

func (c *Contract) GetMatchers() Matchers {
	return c.Matchers
}

// Helper function to convert pcommon.Value to interface{}
func ValueToInterface(value pcommon.Value) interface{} {
	switch value.Type() {
	case pcommon.ValueTypeStr:
		return value.Str()
	case pcommon.ValueTypeInt:
		return value.Int()
	case pcommon.ValueTypeDouble:
		return value.Double()
	case pcommon.ValueTypeBool:
		return value.Bool()
	case pcommon.ValueTypeMap:
		result := make(map[string]interface{})
		value.Map().Range(func(k string, v pcommon.Value) bool {
			result[k] = ValueToInterface(v)
			return true
		})
		return result
	case pcommon.ValueTypeSlice:
		result := make([]interface{}, value.Slice().Len())
		for i := 0; i < value.Slice().Len(); i++ {
			result[i] = ValueToInterface(value.Slice().At(i))
		}
		return result
	default:
		return nil
	}
}
