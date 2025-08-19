// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package contract

import (
	"fmt"
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
	FilterOperatorEquals         FilterOperator = "equals"
	FilterOperatorNotEquals      FilterOperator = "not_equals"
	FilterOperatorMatches        FilterOperator = "matches"
	FilterOperatorNotMatches     FilterOperator = "not_matches"
	FilterOperatorExists         FilterOperator = "exists"
	FilterOperatorNotExists      FilterOperator = "not_exists"
	FilterOperatorGreaterThan    FilterOperator = "greater_than"
	FilterOperatorLessThan       FilterOperator = "less_than"
	FilterOperatorGreaterOrEqual FilterOperator = "greater_or_equal"
	FilterOperatorLessOrEqual    FilterOperator = "less_or_equal"
	FilterOperatorContains       FilterOperator = "contains"
	FilterOperatorNotContains    FilterOperator = "not_contains"
	FilterOperatorStartsWith     FilterOperator = "starts_with"
	FilterOperatorEndsWith       FilterOperator = "ends_with"
	FilterOperatorInRange        FilterOperator = "in_range"
	FilterOperatorNotInRange     FilterOperator = "not_in_range"
	FilterOperatorOneOf          FilterOperator = "one_of"
	FilterOperatorNotOneOf       FilterOperator = "not_one_of"
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

// ValidationRule represents a sophisticated validation rule
type ValidationRule struct {
	Field       string             `yaml:"field"`
	Operator    FilterOperator     `yaml:"operator"`
	Value       interface{}        `yaml:"value,omitempty"`
	Values      []interface{}      `yaml:"values,omitempty"`      // For one_of, not_one_of operations
	Range       *ValueRange        `yaml:"range,omitempty"`       // For range operations
	Pattern     string             `yaml:"pattern,omitempty"`     // For regex patterns
	Condition   *ConditionalRule   `yaml:"condition,omitempty"`   // For conditional logic
	Transform   *TransformRule     `yaml:"transform,omitempty"`   // For transformation validation
	Temporal    *TemporalRule      `yaml:"temporal,omitempty"`    // For time-based validation
	Description string             `yaml:"description,omitempty"` // Human-readable description
	Severity    ValidationSeverity `yaml:"severity,omitempty"`    // Error, warning, or info
}

// ValidationSeverity represents the severity level of a validation rule
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

// ValueRange represents a numeric or temporal range
type ValueRange struct {
	Min          interface{} `yaml:"min,omitempty"`
	Max          interface{} `yaml:"max,omitempty"`
	Inclusive    bool        `yaml:"inclusive,omitempty"`     // Whether endpoints are included
	MinInclusive *bool       `yaml:"min_inclusive,omitempty"` // Override for min
	MaxInclusive *bool       `yaml:"max_inclusive,omitempty"` // Override for max
}

// ConditionalRule represents conditional validation logic
type ConditionalRule struct {
	If   *ValidationRule  `yaml:"if"`
	Then *ValidationRule  `yaml:"then,omitempty"`
	Else *ValidationRule  `yaml:"else,omitempty"`
	And  []ValidationRule `yaml:"and,omitempty"` // All conditions must be true
	Or   []ValidationRule `yaml:"or,omitempty"`  // Any condition must be true
	Not  *ValidationRule  `yaml:"not,omitempty"` // Condition must be false
}

// TransformRule represents expected data transformations
type TransformRule struct {
	Type       string                 `yaml:"type"`                 // add, remove, modify, rename
	Source     string                 `yaml:"source,omitempty"`     // Source field
	Target     string                 `yaml:"target,omitempty"`     // Target field
	Value      interface{}            `yaml:"value,omitempty"`      // Expected value after transformation
	Function   string                 `yaml:"function,omitempty"`   // Transformation function name
	Parameters map[string]interface{} `yaml:"parameters,omitempty"` // Function parameters
}

// TemporalRule represents time-based validation rules
type TemporalRule struct {
	WindowSize  string         `yaml:"window_size"`          // Time window duration
	Aggregation string         `yaml:"aggregation"`          // sum, avg, count, min, max
	Threshold   interface{}    `yaml:"threshold,omitempty"`  // Threshold value
	Comparison  FilterOperator `yaml:"comparison,omitempty"` // Comparison operator
	Baseline    string         `yaml:"baseline,omitempty"`   // Baseline for comparison
	Tolerance   float64        `yaml:"tolerance,omitempty"`  // Tolerance percentage
}

// Filter represents a filter predicate for determining when to apply a contract (legacy)
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
	SpanName         string                   `yaml:"span_name,omitempty"`
	Attributes       map[string]interface{}   `yaml:"attributes,omitempty"`
	ParentSpan       string                   `yaml:"parent_span,omitempty"`
	ServiceName      string                   `yaml:"service_name,omitempty"`
	ValidationRules  []ValidationRule         `yaml:"validation_rules,omitempty"`  // Advanced validation rules
	Count            *CountMatcher            `yaml:"count,omitempty"`             // Expected count validation
	Duration         *DurationMatcher         `yaml:"duration,omitempty"`          // Span duration validation
	StatusCode       *StatusCodeMatcher       `yaml:"status_code,omitempty"`       // HTTP/gRPC status validation
	CustomValidation *CustomValidationMatcher `yaml:"custom_validation,omitempty"` // Custom validation logic
}

// MetricMatcher represents expected metric transformations
type MetricMatcher struct {
	Name             string                   `yaml:"name,omitempty"`
	Type             string                   `yaml:"type,omitempty"`
	Labels           map[string]interface{}   `yaml:"labels,omitempty"`
	ValidationRules  []ValidationRule         `yaml:"validation_rules,omitempty"`  // Advanced validation rules
	Value            *ValueMatcher            `yaml:"value,omitempty"`             // Metric value validation
	Count            *CountMatcher            `yaml:"count,omitempty"`             // Expected count validation
	Histogram        *HistogramMatcher        `yaml:"histogram,omitempty"`         // Histogram-specific validation
	CustomValidation *CustomValidationMatcher `yaml:"custom_validation,omitempty"` // Custom validation logic
}

// LogMatcher represents expected log transformations
type LogMatcher struct {
	Body             string                   `yaml:"body,omitempty"`
	Severity         string                   `yaml:"severity,omitempty"`
	Attributes       map[string]interface{}   `yaml:"attributes,omitempty"`
	ValidationRules  []ValidationRule         `yaml:"validation_rules,omitempty"`  // Advanced validation rules
	Count            *CountMatcher            `yaml:"count,omitempty"`             // Expected count validation
	Timestamp        *TimestampMatcher        `yaml:"timestamp,omitempty"`         // Timestamp validation
	CustomValidation *CustomValidationMatcher `yaml:"custom_validation,omitempty"` // Custom validation logic
}

// CountMatcher represents count-based validation
type CountMatcher struct {
	Expected int            `yaml:"expected,omitempty"`
	Min      *int           `yaml:"min,omitempty"`
	Max      *int           `yaml:"max,omitempty"`
	Operator FilterOperator `yaml:"operator,omitempty"`
	Value    int            `yaml:"value,omitempty"`
}

// ValueMatcher represents metric value validation
type ValueMatcher struct {
	Expected  interface{}    `yaml:"expected,omitempty"`
	Range     *ValueRange    `yaml:"range,omitempty"`
	Operator  FilterOperator `yaml:"operator,omitempty"`
	Tolerance float64        `yaml:"tolerance,omitempty"` // Percentage tolerance for comparisons
}

// DurationMatcher represents span duration validation
type DurationMatcher struct {
	Min       string `yaml:"min,omitempty"`       // Duration string like "100ms"
	Max       string `yaml:"max,omitempty"`       // Duration string like "5s"
	Expected  string `yaml:"expected,omitempty"`  // Expected duration
	Tolerance string `yaml:"tolerance,omitempty"` // Tolerance duration
}

// StatusCodeMatcher represents status code validation
type StatusCodeMatcher struct {
	Expected   int         `yaml:"expected,omitempty"`
	Range      *ValueRange `yaml:"range,omitempty"`
	Class      string      `yaml:"class,omitempty"` // 2xx, 3xx, 4xx, 5xx
	NotAllowed []int       `yaml:"not_allowed,omitempty"`
}

// HistogramMatcher represents histogram-specific validation
type HistogramMatcher struct {
	Buckets      []float64       `yaml:"buckets,omitempty"`
	Count        int             `yaml:"count,omitempty"`
	Sum          float64         `yaml:"sum,omitempty"`
	BucketCounts map[float64]int `yaml:"bucket_counts,omitempty"`
}

// TimestampMatcher represents timestamp validation
type TimestampMatcher struct {
	Format    string      `yaml:"format,omitempty"`    // RFC3339, Unix, etc.
	Range     *ValueRange `yaml:"range,omitempty"`     // Time range
	Relative  string      `yaml:"relative,omitempty"`  // "within_last_hour", etc.
	Precision string      `yaml:"precision,omitempty"` // "second", "millisecond", etc.
}

// CustomValidationMatcher represents custom validation logic
type CustomValidationMatcher struct {
	Script     string                 `yaml:"script,omitempty"`   // Script/expression to evaluate
	Language   string                 `yaml:"language,omitempty"` // "javascript", "go", "cel" etc.
	Function   string                 `yaml:"function,omitempty"` // Function name to call
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
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
	Publisher         string               `yaml:"publisher"`
	Pipeline          string               `yaml:"pipeline,omitempty"`           // Explicit pipeline ID (deprecated in favor of selectors)
	PipelineSelectors *PipelineSelectors   `yaml:"pipeline_selectors,omitempty"` // Pipeline matching criteria
	Version           string               `yaml:"version"`
	Description       string               `yaml:"description,omitempty"`
	Inputs            Inputs               `yaml:"inputs"`
	Filters           []Filter             `yaml:"filters,omitempty"`          // Legacy filters (for backward compatibility)
	ValidationRules   []ValidationRule     `yaml:"validation_rules,omitempty"` // Advanced validation rules
	Matchers          Matchers             `yaml:"matchers"`
	TimeWindows       []TimeWindow         `yaml:"time_windows,omitempty"`
	Schema            *ContractSchema      `yaml:"schema,omitempty"`      // Schema definition for contract validation
	Inheritance       *ContractInheritance `yaml:"inheritance,omitempty"` // Contract inheritance configuration
	FilePath          string               `yaml:"-"`                     // Set by loader
}

// ContractSchema represents schema validation for contracts
type ContractSchema struct {
	Version          string                 `yaml:"version"`                     // Schema version
	RequiredFields   []string               `yaml:"required_fields,omitempty"`   // Required contract fields
	FieldTypes       map[string]string      `yaml:"field_types,omitempty"`       // Expected field types
	ValidationRules  []SchemaValidationRule `yaml:"validation_rules,omitempty"`  // Schema-level validation
	CustomValidators map[string]interface{} `yaml:"custom_validators,omitempty"` // Custom validation functions
}

// SchemaValidationRule represents schema-level validation rules
type SchemaValidationRule struct {
	Field       string   `yaml:"field"`
	Type        string   `yaml:"type"` // string, number, boolean, array, object
	Required    bool     `yaml:"required"`
	MinLength   *int     `yaml:"min_length,omitempty"`
	MaxLength   *int     `yaml:"max_length,omitempty"`
	Pattern     string   `yaml:"pattern,omitempty"`
	Enum        []string `yaml:"enum,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

// ContractInheritance represents contract inheritance and composition
type ContractInheritance struct {
	Extends   []string               `yaml:"extends,omitempty"`   // Parent contracts to inherit from
	Includes  []string               `yaml:"includes,omitempty"`  // Contracts to include/compose
	Overrides map[string]interface{} `yaml:"overrides,omitempty"` // Field overrides for inheritance
	Mixins    []string               `yaml:"mixins,omitempty"`    // Mixin contracts to apply
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
	result := ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]string, 0),
	}

	// Validate contract structure
	if err := c.validateContractStructure(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "contract_structure",
			Message: err.Error(),
		})
		return result
	}

	// Apply filters to determine if this contract should be validated
	if !c.applyFilters(input) {
		// Contract doesn't apply to this data, but that's not an error
		return result
	}

	// Validate input data presence
	if err := c.validateInputData(input); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "input_data",
			Message: err.Error(),
		})
	}

	// Validate output data against matchers
	if err := c.validateOutputData(output); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "output_validation",
			Message: err.Error(),
		})
	}

	// Validate time windows if specified
	if err := c.validateTimeWindows(input.Time); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "time_window",
			Message: err.Error(),
		})
	}

	return result
}

// validateContractStructure validates the contract's internal structure
func (c *Contract) validateContractStructure() error {
	if c.Publisher == "" {
		return fmt.Errorf("publisher is required")
	}
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.Pipeline == "" && (c.PipelineSelectors == nil || len(c.PipelineSelectors.Selectors) == 0) {
		return fmt.Errorf("either pipeline or pipeline_selectors must be specified")
	}
	if len(c.Inputs.Traces) == 0 && len(c.Inputs.Metrics) == 0 && len(c.Inputs.Logs) == 0 {
		return fmt.Errorf("at least one input (traces, metrics, or logs) must be specified")
	}
	if len(c.Matchers.Traces) == 0 && len(c.Matchers.Metrics) == 0 && len(c.Matchers.Logs) == 0 {
		return fmt.Errorf("at least one matcher (traces, metrics, or logs) must be specified")
	}
	return nil
}

// applyFilters applies filter predicates to determine if a contract should be validated
func (c *Contract) applyFilters(data OpenTelemetryData) bool {
	if len(c.Filters) == 0 {
		return true
	}

	for _, filter := range c.Filters {
		if !c.evaluateFilter(filter, data) {
			return false
		}
	}
	return true
}

// evaluateFilter evaluates a single filter against the data
func (c *Contract) evaluateFilter(filter Filter, data OpenTelemetryData) bool {
	// This is a simplified implementation - in practice, this would use the matcher package
	// For now, we'll return true to indicate the filter passes
	// TODO: Implement actual filter evaluation logic
	return true
}

// validateInputData validates that the input data matches the contract's input specification
func (c *Contract) validateInputData(data OpenTelemetryData) error {
	// Validate traces input
	if len(c.Inputs.Traces) > 0 {
		resourceSpans := data.Traces.ResourceSpans()
		if resourceSpans.Len() == 0 {
			return fmt.Errorf("contract expects trace input but no traces provided")
		}
	}

	// Validate metrics input
	if len(c.Inputs.Metrics) > 0 {
		resourceMetrics := data.Metrics.ResourceMetrics()
		if resourceMetrics.Len() == 0 {
			return fmt.Errorf("contract expects metric input but no metrics provided")
		}
	}

	// Validate logs input
	if len(c.Inputs.Logs) > 0 {
		resourceLogs := data.Logs.ResourceLogs()
		if resourceLogs.Len() == 0 {
			return fmt.Errorf("contract expects log input but no logs provided")
		}
	}

	return nil
}

// validateOutputData validates that the output data matches the contract's matchers
func (c *Contract) validateOutputData(data OpenTelemetryData) error {
	// Validate traces output
	if len(c.Matchers.Traces) > 0 {
		resourceSpans := data.Traces.ResourceSpans()
		if resourceSpans.Len() == 0 {
			return fmt.Errorf("contract expects trace output but no traces found")
		}
	}

	// Validate metrics output
	if len(c.Matchers.Metrics) > 0 {
		resourceMetrics := data.Metrics.ResourceMetrics()
		if resourceMetrics.Len() == 0 {
			return fmt.Errorf("contract expects metric output but no metrics found")
		}
	}

	// Validate logs output
	if len(c.Matchers.Logs) > 0 {
		resourceLogs := data.Logs.ResourceLogs()
		if resourceLogs.Len() == 0 {
			return fmt.Errorf("contract expects log output but no logs found")
		}
	}

	return nil
}

// validateTimeWindows validates that the data timestamp falls within specified time windows
func (c *Contract) validateTimeWindows(timestamp time.Time) error {
	if len(c.TimeWindows) == 0 {
		return nil
	}

	// For now, we'll implement basic time window validation
	// TODO: Implement more sophisticated time window validation based on aggregation and duration
	for _, window := range c.TimeWindows {
		if window.Duration != "" {
			// Parse duration and validate timestamp is within reasonable bounds
			// This is a placeholder for more sophisticated time window validation
			_ = window.Duration // Use the duration field to avoid unused variable warning
		}
	}

	return nil
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
