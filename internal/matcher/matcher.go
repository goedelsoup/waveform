// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package matcher

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/goedelsoup/waveform/internal/contract"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// Matcher handles validation of OpenTelemetry data against contracts
type Matcher struct {
	ignoreTimestamps bool
	timeTolerance    time.Duration
}

// NewMatcher creates a new matcher instance
func NewMatcher() *Matcher {
	return &Matcher{
		ignoreTimestamps: true,
		timeTolerance:    1 * time.Second,
	}
}

// SetIgnoreTimestamps sets whether timestamps should be ignored during matching
func (m *Matcher) SetIgnoreTimestamps(ignore bool) {
	m.ignoreTimestamps = ignore
}

// SetTimeTolerance sets the tolerance for timestamp comparisons
func (m *Matcher) SetTimeTolerance(tolerance time.Duration) {
	m.timeTolerance = tolerance
}

// Validate validates output data against contract expectations
func (m *Matcher) Validate(contractDef *contract.Contract, input, output contract.OpenTelemetryData) contract.ValidationResult {
	result := contract.ValidationResult{
		Valid:    true,
		Errors:   make([]contract.ValidationError, 0),
		Warnings: make([]string, 0),
	}

	// Apply filters to determine if this contract should be validated
	if !m.applyFilters(contractDef.Filters, input) {
		return result
	}

	// Validate traces
	if len(contractDef.Matchers.Traces) > 0 {
		if err := m.validateTraces(contractDef.Matchers.Traces, output.Traces); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, contract.ValidationError{
				Type:       "trace_validation",
				Message:    err.Error(),
				SignalType: contract.SignalTypeTraces,
			})
		}
	}

	// Validate metrics
	if len(contractDef.Matchers.Metrics) > 0 {
		if err := m.validateMetrics(contractDef.Matchers.Metrics, output.Metrics); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, contract.ValidationError{
				Type:       "metric_validation",
				Message:    err.Error(),
				SignalType: contract.SignalTypeMetrics,
			})
		}
	}

	// Validate logs
	if len(contractDef.Matchers.Logs) > 0 {
		if err := m.validateLogs(contractDef.Matchers.Logs, output.Logs); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, contract.ValidationError{
				Type:       "log_validation",
				Message:    err.Error(),
				SignalType: contract.SignalTypeLogs,
			})
		}
	}

	return result
}

// applyFilters applies filter predicates to determine if a contract should be validated
func (m *Matcher) applyFilters(filters []contract.Filter, data contract.OpenTelemetryData) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		if !m.evaluateFilter(filter, data) {
			return false
		}
	}
	return true
}

// evaluateFilter evaluates a single filter against the data
func (m *Matcher) evaluateFilter(filter contract.Filter, data contract.OpenTelemetryData) bool {
	// Extract the field value based on the field path
	fieldValue := m.extractFieldValue(filter.Field, data)

	switch filter.Operator {
	case contract.FilterOperatorEquals:
		return m.compareValues(fieldValue, filter.Value, "equals")
	case contract.FilterOperatorNotEquals:
		return !m.compareValues(fieldValue, filter.Value, "equals")
	case contract.FilterOperatorMatches:
		return m.matchesPattern(fieldValue, filter.Value)
	case contract.FilterOperatorExists:
		return fieldValue != nil
	case contract.FilterOperatorNotExists:
		return fieldValue == nil
	case contract.FilterOperatorGreaterThan:
		return m.compareValues(fieldValue, filter.Value, "greater_than")
	case contract.FilterOperatorLessThan:
		return m.compareValues(fieldValue, filter.Value, "less_than")
	default:
		return false
	}
}

// extractFieldValue extracts a field value from the data based on a dot-separated path
func (m *Matcher) extractFieldValue(fieldPath string, data contract.OpenTelemetryData) interface{} {
	parts := strings.Split(fieldPath, ".")
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "span":
		return m.extractSpanField(parts[1:], data.Traces)
	case "metric":
		return m.extractMetricField(parts[1:], data.Metrics)
	case "log":
		return m.extractLogField(parts[1:], data.Logs)
	default:
		return nil
	}
}

// extractSpanField extracts a field from span data
func (m *Matcher) extractSpanField(parts []string, traces ptrace.Traces) interface{} {
	if traces.ResourceSpans().Len() == 0 {
		return nil
	}

	// For simplicity, we'll look at the first span
	resourceSpans := traces.ResourceSpans().At(0)
	if resourceSpans.ScopeSpans().Len() == 0 {
		return nil
	}

	scopeSpans := resourceSpans.ScopeSpans().At(0)
	if scopeSpans.Spans().Len() == 0 {
		return nil
	}

	span := scopeSpans.Spans().At(0)

	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "name":
		return span.Name()
	case "service", "service.name":
		if val, ok := resourceSpans.Resource().Attributes().Get("service.name"); ok {
			return val.Str()
		}
		return ""
	case "attributes":
		if len(parts) > 1 {
			if val, ok := span.Attributes().Get(parts[1]); ok {
				return val.AsString()
			}
		}
		return nil
	default:
		return nil
	}
}

// extractMetricField extracts a field from metric data
func (m *Matcher) extractMetricField(parts []string, metrics pmetric.Metrics) interface{} {
	if metrics.ResourceMetrics().Len() == 0 {
		return nil
	}

	resourceMetrics := metrics.ResourceMetrics().At(0)
	if resourceMetrics.ScopeMetrics().Len() == 0 {
		return nil
	}

	scopeMetrics := resourceMetrics.ScopeMetrics().At(0)
	if scopeMetrics.Metrics().Len() == 0 {
		return nil
	}

	metric := scopeMetrics.Metrics().At(0)

	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "name":
		return metric.Name()
	case "type":
		switch metric.Type() {
		case pmetric.MetricTypeGauge:
			return "gauge"
		case pmetric.MetricTypeSum:
			return "sum"
		case pmetric.MetricTypeHistogram:
			return "histogram"
		default:
			return "unknown"
		}
	case "labels":
		if len(parts) > 1 {
			// For simplicity, we'll look at the first data point
			switch metric.Type() {
			case pmetric.MetricTypeGauge:
				if metric.Gauge().DataPoints().Len() > 0 {
					if val, ok := metric.Gauge().DataPoints().At(0).Attributes().Get(parts[1]); ok {
						return val.AsString()
					}
				}
			case pmetric.MetricTypeSum:
				if metric.Sum().DataPoints().Len() > 0 {
					if val, ok := metric.Sum().DataPoints().At(0).Attributes().Get(parts[1]); ok {
						return val.AsString()
					}
				}
			}
		}
		return nil
	default:
		return nil
	}
}

// extractLogField extracts a field from log data
func (m *Matcher) extractLogField(parts []string, logs plog.Logs) interface{} {
	if logs.ResourceLogs().Len() == 0 {
		return nil
	}

	resourceLogs := logs.ResourceLogs().At(0)
	if resourceLogs.ScopeLogs().Len() == 0 {
		return nil
	}

	scopeLogs := resourceLogs.ScopeLogs().At(0)
	if scopeLogs.LogRecords().Len() == 0 {
		return nil
	}

	logRecord := scopeLogs.LogRecords().At(0)

	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "body":
		return logRecord.Body().AsString()
	case "severity":
		return logRecord.SeverityText()
	case "attributes":
		if len(parts) > 1 {
			if val, ok := logRecord.Attributes().Get(parts[1]); ok {
				return val.AsString()
			}
		}
		return nil
	default:
		return nil
	}
}

// compareValues compares two values for equality or ordering
func (m *Matcher) compareValues(a, b interface{}, operation string) bool {
	if a == nil || b == nil {
		return a == b
	}

	// Try to convert to comparable types
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			switch operation {
			case "equals":
				return av == bv
			case "greater_than":
				return av > bv
			case "less_than":
				return av < bv
			}
		}
	case int, int64:
		var avInt int64
		switch v := av.(type) {
		case int:
			avInt = int64(v)
		case int64:
			avInt = v
		}

		switch bv := b.(type) {
		case int:
			switch operation {
			case "equals":
				return avInt == int64(bv)
			case "greater_than":
				return avInt > int64(bv)
			case "less_than":
				return avInt < int64(bv)
			}
		case int64:
			switch operation {
			case "equals":
				return avInt == bv
			case "greater_than":
				return avInt > bv
			case "less_than":
				return avInt < bv
			}
		}
	case float64:
		if bv, ok := b.(float64); ok {
			switch operation {
			case "equals":
				return av == bv
			case "greater_than":
				return av > bv
			case "less_than":
				return av < bv
			}
		}
	}

	return false
}

// matchesPattern checks if a value matches a regex pattern
func (m *Matcher) matchesPattern(value interface{}, pattern interface{}) bool {
	if value == nil || pattern == nil {
		return false
	}

	valueStr, ok := value.(string)
	if !ok {
		valueStr = fmt.Sprintf("%v", value)
	}

	patternStr, ok := pattern.(string)
	if !ok {
		return false
	}

	regex, err := regexp.Compile(patternStr)
	if err != nil {
		return false
	}

	return regex.MatchString(valueStr)
}

// validateTraces validates trace data against matchers
func (m *Matcher) validateTraces(matchers []contract.TraceMatcher, traces ptrace.Traces) error {
	if traces.ResourceSpans().Len() == 0 {
		return fmt.Errorf("no traces found in output")
	}

	for i, matcher := range matchers {
		if err := m.validateTrace(matcher, traces); err != nil {
			return fmt.Errorf("trace matcher %d failed: %w", i, err)
		}
	}

	return nil
}

// validateTrace validates a single trace against a matcher
func (m *Matcher) validateTrace(matcher contract.TraceMatcher, traces ptrace.Traces) error {
	// For simplicity, we'll validate against the first span
	resourceSpans := traces.ResourceSpans().At(0)
	scopeSpans := resourceSpans.ScopeSpans().At(0)
	span := scopeSpans.Spans().At(0)

	// Validate span name
	if matcher.SpanName != "" && span.Name() != matcher.SpanName {
		return fmt.Errorf("span name mismatch: expected %s, got %s", matcher.SpanName, span.Name())
	}

	// Validate service name
	if matcher.ServiceName != "" {
		if serviceName, ok := resourceSpans.Resource().Attributes().Get("service.name"); ok {
			if serviceName.Str() != matcher.ServiceName {
				return fmt.Errorf("service name mismatch: expected %s, got %s", matcher.ServiceName, serviceName.Str())
			}
		} else {
			return fmt.Errorf("service name not found in resource attributes")
		}
	}

	// Validate attributes
	for key, expectedValue := range matcher.Attributes {
		if strings.HasPrefix(key, "!") {
			// Negation - field should not exist
			fieldName := strings.TrimPrefix(key, "!")
			if span.Attributes().Len() > 0 {
				if _, exists := span.Attributes().Get(fieldName); exists {
					return fmt.Errorf("attribute %s should not exist", fieldName)
				}
			}
		} else {
			// Field should exist and match
			if actualValue, ok := span.Attributes().Get(key); ok {
				if actualValue.Str() != expectedValue {
					return fmt.Errorf("attribute %s mismatch: expected %v, got %s", key, expectedValue, actualValue.Str())
				}
			} else {
				return fmt.Errorf("attribute %s not found", key)
			}
		}
	}

	return nil
}

// validateMetrics validates metric data against matchers
func (m *Matcher) validateMetrics(matchers []contract.MetricMatcher, metrics pmetric.Metrics) error {
	if metrics.ResourceMetrics().Len() == 0 {
		return fmt.Errorf("no metrics found in output")
	}

	for i, matcher := range matchers {
		if err := m.validateMetric(matcher, metrics); err != nil {
			return fmt.Errorf("metric matcher %d failed: %w", i, err)
		}
	}

	return nil
}

// validateMetric validates a single metric against a matcher
func (m *Matcher) validateMetric(matcher contract.MetricMatcher, metrics pmetric.Metrics) error {
	resourceMetrics := metrics.ResourceMetrics().At(0)
	scopeMetrics := resourceMetrics.ScopeMetrics().At(0)
	metric := scopeMetrics.Metrics().At(0)

	// Validate metric name
	if matcher.Name != "" && metric.Name() != matcher.Name {
		return fmt.Errorf("metric name mismatch: expected %s, got %s", matcher.Name, metric.Name())
	}

	// Validate metric type
	if matcher.Type != "" {
		actualType := ""
		switch metric.Type() {
		case pmetric.MetricTypeGauge:
			actualType = "gauge"
		case pmetric.MetricTypeSum:
			actualType = "sum"
		case pmetric.MetricTypeHistogram:
			actualType = "histogram"
		}
		if actualType != matcher.Type {
			return fmt.Errorf("metric type mismatch: expected %s, got %s", matcher.Type, actualType)
		}
	}

	// Validate labels
	for key, expectedValue := range matcher.Labels {
		var actualValue pcommon.Value
		var found bool
		switch metric.Type() {
		case pmetric.MetricTypeGauge:
			if metric.Gauge().DataPoints().Len() > 0 {
				actualValue, found = metric.Gauge().DataPoints().At(0).Attributes().Get(key)
			}
		case pmetric.MetricTypeSum:
			if metric.Sum().DataPoints().Len() > 0 {
				actualValue, found = metric.Sum().DataPoints().At(0).Attributes().Get(key)
			}
		}

		if !found {
			return fmt.Errorf("label %s not found", key)
		}
		if actualValue.Str() != expectedValue {
			return fmt.Errorf("label %s mismatch: expected %v, got %s", key, expectedValue, actualValue.Str())
		}
	}

	return nil
}

// validateLogs validates log data against matchers
func (m *Matcher) validateLogs(matchers []contract.LogMatcher, logs plog.Logs) error {
	if logs.ResourceLogs().Len() == 0 {
		return fmt.Errorf("no logs found in output")
	}

	for i, matcher := range matchers {
		if err := m.validateLog(matcher, logs); err != nil {
			return fmt.Errorf("log matcher %d failed: %w", i, err)
		}
	}

	return nil
}

// validateLog validates a single log against a matcher
func (m *Matcher) validateLog(matcher contract.LogMatcher, logs plog.Logs) error {
	resourceLogs := logs.ResourceLogs().At(0)
	scopeLogs := resourceLogs.ScopeLogs().At(0)
	logRecord := scopeLogs.LogRecords().At(0)

	// Validate log body
	if matcher.Body != "" && logRecord.Body().AsString() != matcher.Body {
		return fmt.Errorf("log body mismatch: expected %s, got %s", matcher.Body, logRecord.Body().AsString())
	}

	// Validate severity
	if matcher.Severity != "" && logRecord.SeverityText() != matcher.Severity {
		return fmt.Errorf("log severity mismatch: expected %s, got %s", matcher.Severity, logRecord.SeverityText())
	}

	// Validate attributes
	for key, expectedValue := range matcher.Attributes {
		if actualValue, ok := logRecord.Attributes().Get(key); ok {
			if actualValue.Str() != expectedValue {
				return fmt.Errorf("log attribute %s mismatch: expected %v, got %s", key, expectedValue, actualValue.Str())
			}
		} else {
			return fmt.Errorf("log attribute %s not found", key)
		}
	}

	return nil
}
