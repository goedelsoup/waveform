// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package matcher

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goedelsoup/waveform/internal/contract"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// AdvancedMatcher handles sophisticated validation rules and patterns
type AdvancedMatcher struct {
	logger           *zap.Logger
	ignoreTimestamps bool
	timeTolerance    time.Duration
	customValidators map[string]func(interface{}, map[string]interface{}) bool
}

// NewAdvancedMatcher creates a new advanced matcher instance
func NewAdvancedMatcher(logger *zap.Logger) *AdvancedMatcher {
	return &AdvancedMatcher{
		logger:           logger,
		ignoreTimestamps: true,
		timeTolerance:    1 * time.Second,
		customValidators: make(map[string]func(interface{}, map[string]interface{}) bool),
	}
}

// RegisterCustomValidator registers a custom validation function
func (am *AdvancedMatcher) RegisterCustomValidator(name string, validator func(interface{}, map[string]interface{}) bool) {
	am.customValidators[name] = validator
}

// ValidateRule validates a single validation rule against data
func (am *AdvancedMatcher) ValidateRule(rule contract.ValidationRule, data contract.OpenTelemetryData) error {
	// Extract the field value
	fieldValue := am.extractFieldValue(rule.Field, data)

	// Handle conditional logic first
	if rule.Condition != nil {
		return am.validateConditionalRule(*rule.Condition, data)
	}

	// Handle transformation validation
	if rule.Transform != nil {
		return am.validateTransformRule(*rule.Transform, fieldValue, data)
	}

	// Handle temporal validation
	if rule.Temporal != nil {
		return am.validateTemporalRule(*rule.Temporal, fieldValue, data)
	}

	// Handle basic validation based on operator
	return am.validateBasicRule(rule, fieldValue)
}

// validateBasicRule validates basic rules with operators
func (am *AdvancedMatcher) validateBasicRule(rule contract.ValidationRule, fieldValue interface{}) error {
	switch rule.Operator {
	case contract.FilterOperatorEquals:
		if !am.compareValues(fieldValue, rule.Value, "equals") {
			return fmt.Errorf("field %s: expected %v, got %v", rule.Field, rule.Value, fieldValue)
		}
	case contract.FilterOperatorNotEquals:
		if am.compareValues(fieldValue, rule.Value, "equals") {
			return fmt.Errorf("field %s: should not equal %v", rule.Field, rule.Value)
		}
	case contract.FilterOperatorMatches:
		pattern := rule.Pattern
		if pattern == "" && rule.Value != nil {
			pattern = fmt.Sprintf("%v", rule.Value)
		}
		if !am.matchesPattern(fieldValue, pattern) {
			return fmt.Errorf("field %s: value %v does not match pattern %s", rule.Field, fieldValue, pattern)
		}
	case contract.FilterOperatorNotMatches:
		pattern := rule.Pattern
		if pattern == "" && rule.Value != nil {
			pattern = fmt.Sprintf("%v", rule.Value)
		}
		if am.matchesPattern(fieldValue, pattern) {
			return fmt.Errorf("field %s: value %v should not match pattern %s", rule.Field, fieldValue, pattern)
		}
	case contract.FilterOperatorExists:
		if fieldValue == nil {
			return fmt.Errorf("field %s: should exist", rule.Field)
		}
	case contract.FilterOperatorNotExists:
		if fieldValue != nil {
			return fmt.Errorf("field %s: should not exist", rule.Field)
		}
	case contract.FilterOperatorGreaterThan:
		if !am.compareValues(fieldValue, rule.Value, "greater_than") {
			return fmt.Errorf("field %s: %v should be greater than %v", rule.Field, fieldValue, rule.Value)
		}
	case contract.FilterOperatorLessThan:
		if !am.compareValues(fieldValue, rule.Value, "less_than") {
			return fmt.Errorf("field %s: %v should be less than %v", rule.Field, fieldValue, rule.Value)
		}
	case contract.FilterOperatorGreaterOrEqual:
		if !am.compareValues(fieldValue, rule.Value, "greater_or_equal") {
			return fmt.Errorf("field %s: %v should be greater than or equal to %v", rule.Field, fieldValue, rule.Value)
		}
	case contract.FilterOperatorLessOrEqual:
		if !am.compareValues(fieldValue, rule.Value, "less_or_equal") {
			return fmt.Errorf("field %s: %v should be less than or equal to %v", rule.Field, fieldValue, rule.Value)
		}
	case contract.FilterOperatorContains:
		if !am.stringContains(fieldValue, rule.Value) {
			return fmt.Errorf("field %s: %v should contain %v", rule.Field, fieldValue, rule.Value)
		}
	case contract.FilterOperatorNotContains:
		if am.stringContains(fieldValue, rule.Value) {
			return fmt.Errorf("field %s: %v should not contain %v", rule.Field, fieldValue, rule.Value)
		}
	case contract.FilterOperatorStartsWith:
		if !am.stringStartsWith(fieldValue, rule.Value) {
			return fmt.Errorf("field %s: %v should start with %v", rule.Field, fieldValue, rule.Value)
		}
	case contract.FilterOperatorEndsWith:
		if !am.stringEndsWith(fieldValue, rule.Value) {
			return fmt.Errorf("field %s: %v should end with %v", rule.Field, fieldValue, rule.Value)
		}
	case contract.FilterOperatorInRange:
		if rule.Range == nil {
			return fmt.Errorf("field %s: range not specified for in_range operator", rule.Field)
		}
		if !am.inRange(fieldValue, *rule.Range) {
			return fmt.Errorf("field %s: %v not in range [%v, %v]", rule.Field, fieldValue, rule.Range.Min, rule.Range.Max)
		}
	case contract.FilterOperatorNotInRange:
		if rule.Range == nil {
			return fmt.Errorf("field %s: range not specified for not_in_range operator", rule.Field)
		}
		if am.inRange(fieldValue, *rule.Range) {
			return fmt.Errorf("field %s: %v should not be in range [%v, %v]", rule.Field, fieldValue, rule.Range.Min, rule.Range.Max)
		}
	case contract.FilterOperatorOneOf:
		if !am.oneOf(fieldValue, rule.Values) {
			return fmt.Errorf("field %s: %v should be one of %v", rule.Field, fieldValue, rule.Values)
		}
	case contract.FilterOperatorNotOneOf:
		if am.oneOf(fieldValue, rule.Values) {
			return fmt.Errorf("field %s: %v should not be one of %v", rule.Field, fieldValue, rule.Values)
		}
	default:
		return fmt.Errorf("unsupported operator: %s", rule.Operator)
	}

	return nil
}

// validateConditionalRule validates conditional logic
func (am *AdvancedMatcher) validateConditionalRule(condition contract.ConditionalRule, data contract.OpenTelemetryData) error {
	// Handle If-Then-Else logic
	if condition.If != nil {
		ifResult := am.ValidateRule(*condition.If, data)
		if ifResult == nil { // If condition is true
			if condition.Then != nil {
				return am.ValidateRule(*condition.Then, data)
			}
		} else { // If condition is false
			if condition.Else != nil {
				return am.ValidateRule(*condition.Else, data)
			}
		}
	}

	// Handle AND logic - all conditions must be true
	if len(condition.And) > 0 {
		for i, rule := range condition.And {
			if err := am.ValidateRule(rule, data); err != nil {
				return fmt.Errorf("AND condition %d failed: %w", i, err)
			}
		}
	}

	// Handle OR logic - at least one condition must be true
	if len(condition.Or) > 0 {
		var errors []string
		for _, rule := range condition.Or {
			if err := am.ValidateRule(rule, data); err == nil {
				return nil // At least one passed
			} else {
				errors = append(errors, err.Error())
			}
		}
		return fmt.Errorf("all OR conditions failed: %v", errors)
	}

	// Handle NOT logic - condition must be false
	if condition.Not != nil {
		if err := am.ValidateRule(*condition.Not, data); err == nil {
			return fmt.Errorf("NOT condition should have failed")
		}
	}

	return nil
}

// validateTransformRule validates transformation expectations
func (am *AdvancedMatcher) validateTransformRule(transform contract.TransformRule, fieldValue interface{}, data contract.OpenTelemetryData) error {
	switch transform.Type {
	case "add":
		// Validate that a field was added
		if fieldValue == nil {
			return fmt.Errorf("expected field %s to be added", transform.Target)
		}
		if transform.Value != nil && !am.compareValues(fieldValue, transform.Value, "equals") {
			return fmt.Errorf("added field %s: expected %v, got %v", transform.Target, transform.Value, fieldValue)
		}
	case "remove":
		// Validate that a field was removed
		if fieldValue != nil {
			return fmt.Errorf("expected field %s to be removed", transform.Source)
		}
	case "modify":
		// Validate that a field was modified to expected value
		if transform.Value != nil && !am.compareValues(fieldValue, transform.Value, "equals") {
			return fmt.Errorf("modified field %s: expected %v, got %v", transform.Target, transform.Value, fieldValue)
		}
	case "rename":
		// Validate that source field was removed and target field exists
		sourceValue := am.extractFieldValue(transform.Source, data)
		if sourceValue != nil {
			return fmt.Errorf("expected source field %s to be removed after rename", transform.Source)
		}
		if fieldValue == nil {
			return fmt.Errorf("expected target field %s to exist after rename", transform.Target)
		}
	default:
		return fmt.Errorf("unsupported transformation type: %s", transform.Type)
	}

	return nil
}

// validateTemporalRule validates time-based rules
func (am *AdvancedMatcher) validateTemporalRule(temporal contract.TemporalRule, fieldValue interface{}, data contract.OpenTelemetryData) error {
	// Parse window size
	windowDuration, err := time.ParseDuration(temporal.WindowSize)
	if err != nil {
		return fmt.Errorf("invalid window size: %w", err)
	}

	// Get current time and window boundaries
	now := time.Now()
	windowStart := now.Add(-windowDuration)

	// Extract timestamp from data
	var timestamp time.Time
	if data.Time.IsZero() {
		timestamp = now
	} else {
		timestamp = data.Time
	}

	// Check if timestamp is within window
	if timestamp.Before(windowStart) {
		return fmt.Errorf("timestamp %v is outside window [%v, %v]", timestamp, windowStart, now)
	}

	// Validate threshold if specified
	if temporal.Threshold != nil && temporal.Comparison != "" {
		numericValue, err := am.toNumeric(fieldValue)
		if err != nil {
			return fmt.Errorf("cannot convert field value to numeric for temporal validation: %w", err)
		}

		thresholdValue, err := am.toNumeric(temporal.Threshold)
		if err != nil {
			return fmt.Errorf("cannot convert threshold to numeric: %w", err)
		}

		if !am.compareNumeric(numericValue, thresholdValue, string(temporal.Comparison)) {
			return fmt.Errorf("temporal validation failed: %v %s %v", numericValue, temporal.Comparison, thresholdValue)
		}
	}

	return nil
}

// Helper methods for validation operations

func (am *AdvancedMatcher) extractFieldValue(fieldPath string, data contract.OpenTelemetryData) interface{} {
	parts := strings.Split(fieldPath, ".")
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "span":
		return am.extractSpanField(parts[1:], data.Traces)
	case "metric":
		return am.extractMetricField(parts[1:], data.Metrics)
	case "log":
		return am.extractLogField(parts[1:], data.Logs)
	default:
		return nil
	}
}

func (am *AdvancedMatcher) extractSpanField(parts []string, traces ptrace.Traces) interface{} {
	if traces.ResourceSpans().Len() == 0 || len(parts) == 0 {
		return nil
	}

	resourceSpans := traces.ResourceSpans().At(0)
	if resourceSpans.ScopeSpans().Len() == 0 {
		return nil
	}

	scopeSpans := resourceSpans.ScopeSpans().At(0)
	if scopeSpans.Spans().Len() == 0 {
		return nil
	}

	span := scopeSpans.Spans().At(0)

	switch parts[0] {
	case "name":
		return span.Name()
	case "service", "service.name":
		if val, ok := resourceSpans.Resource().Attributes().Get("service.name"); ok {
			return val.Str()
		}
		return nil
	case "duration":
		return span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime()).String()
	case "status":
		return span.Status().Code().String()
	case "attributes":
		if len(parts) > 1 {
			if val, ok := span.Attributes().Get(parts[1]); ok {
				return contract.ValueToInterface(val)
			}
		}
		return nil
	default:
		return nil
	}
}

func (am *AdvancedMatcher) extractMetricField(parts []string, metrics pmetric.Metrics) interface{} {
	if metrics.ResourceMetrics().Len() == 0 || len(parts) == 0 {
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

	switch parts[0] {
	case "name":
		return metric.Name()
	case "type":
		return metric.Type().String()
	case "value":
		// Return the first data point value
		switch metric.Type() {
		case pmetric.MetricTypeGauge:
			if metric.Gauge().DataPoints().Len() > 0 {
				return metric.Gauge().DataPoints().At(0).DoubleValue()
			}
		case pmetric.MetricTypeSum:
			if metric.Sum().DataPoints().Len() > 0 {
				return metric.Sum().DataPoints().At(0).DoubleValue()
			}
		}
		return nil
	case "labels":
		if len(parts) > 1 {
			switch metric.Type() {
			case pmetric.MetricTypeGauge:
				if metric.Gauge().DataPoints().Len() > 0 {
					if val, ok := metric.Gauge().DataPoints().At(0).Attributes().Get(parts[1]); ok {
						return contract.ValueToInterface(val)
					}
				}
			case pmetric.MetricTypeSum:
				if metric.Sum().DataPoints().Len() > 0 {
					if val, ok := metric.Sum().DataPoints().At(0).Attributes().Get(parts[1]); ok {
						return contract.ValueToInterface(val)
					}
				}
			}
		}
		return nil
	default:
		return nil
	}
}

func (am *AdvancedMatcher) extractLogField(parts []string, logs plog.Logs) interface{} {
	if logs.ResourceLogs().Len() == 0 || len(parts) == 0 {
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

	switch parts[0] {
	case "body":
		return logRecord.Body().AsString()
	case "severity":
		return logRecord.SeverityText()
	case "timestamp":
		return logRecord.Timestamp().AsTime()
	case "attributes":
		if len(parts) > 1 {
			if val, ok := logRecord.Attributes().Get(parts[1]); ok {
				return contract.ValueToInterface(val)
			}
		}
		return nil
	default:
		return nil
	}
}

func (am *AdvancedMatcher) compareValues(a, b interface{}, operation string) bool {
	if a == nil || b == nil {
		return a == b
	}

	// Convert to comparable types
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
			case "greater_or_equal":
				return av >= bv
			case "less_or_equal":
				return av <= bv
			}
		}
	case int, int64, float64:
		aNum, _ := am.toNumeric(a)
		bNum, _ := am.toNumeric(b)
		return am.compareNumeric(aNum, bNum, operation)
	}

	return false
}

func (am *AdvancedMatcher) compareNumeric(a, b float64, operation string) bool {
	switch operation {
	case "equals":
		return math.Abs(a-b) < 1e-9
	case "greater_than":
		return a > b
	case "less_than":
		return a < b
	case "greater_or_equal":
		return a >= b
	case "less_or_equal":
		return a <= b
	}
	return false
}

func (am *AdvancedMatcher) toNumeric(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to numeric", value)
	}
}

func (am *AdvancedMatcher) matchesPattern(value interface{}, pattern string) bool {
	if value == nil || pattern == "" {
		return false
	}

	valueStr := fmt.Sprintf("%v", value)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	return regex.MatchString(valueStr)
}

func (am *AdvancedMatcher) stringContains(value, substring interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	subStr := fmt.Sprintf("%v", substring)
	return strings.Contains(valueStr, subStr)
}

func (am *AdvancedMatcher) stringStartsWith(value, prefix interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	prefixStr := fmt.Sprintf("%v", prefix)
	return strings.HasPrefix(valueStr, prefixStr)
}

func (am *AdvancedMatcher) stringEndsWith(value, suffix interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	suffixStr := fmt.Sprintf("%v", suffix)
	return strings.HasSuffix(valueStr, suffixStr)
}

func (am *AdvancedMatcher) inRange(value interface{}, valueRange contract.ValueRange) bool {
	numValue, err := am.toNumeric(value)
	if err != nil {
		return false
	}

	// Check minimum
	if valueRange.Min != nil {
		minValue, err := am.toNumeric(valueRange.Min)
		if err != nil {
			return false
		}

		minInclusive := valueRange.Inclusive
		if valueRange.MinInclusive != nil {
			minInclusive = *valueRange.MinInclusive
		}

		if minInclusive {
			if numValue < minValue {
				return false
			}
		} else {
			if numValue <= minValue {
				return false
			}
		}
	}

	// Check maximum
	if valueRange.Max != nil {
		maxValue, err := am.toNumeric(valueRange.Max)
		if err != nil {
			return false
		}

		maxInclusive := valueRange.Inclusive
		if valueRange.MaxInclusive != nil {
			maxInclusive = *valueRange.MaxInclusive
		}

		if maxInclusive {
			if numValue > maxValue {
				return false
			}
		} else {
			if numValue >= maxValue {
				return false
			}
		}
	}

	return true
}

func (am *AdvancedMatcher) oneOf(value interface{}, values []interface{}) bool {
	for _, v := range values {
		if am.compareValues(value, v, "equals") {
			return true
		}
	}
	return false
}
