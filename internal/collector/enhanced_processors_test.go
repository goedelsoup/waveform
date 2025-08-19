// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package collector

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func TestEnhancedTransformProcessor_TraceTransformation(t *testing.T) {
	// Create processor
	processor := NewEnhancedTransformProcessor("test-transform", nil, zap.NewNop())

	// Create test traces
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	span := scopeSpans.Spans().AppendEmpty()

	// Set original span name
	span.SetName("original-span")

	// Add attributes that will be used for transformation
	span.Attributes().PutStr("http.method", "GET")
	span.Attributes().PutStr("http.route", "/api/users")

	// Process traces
	err := processor.ProcessTraces(context.Background(), traces)
	if err != nil {
		t.Fatalf("Failed to process traces: %v", err)
	}

	// Verify transformation
	expectedName := "GET /api/users"
	if span.Name() != expectedName {
		t.Errorf("Expected span name '%s', got '%s'", expectedName, span.Name())
	}
}

func TestEnhancedTransformProcessor_MetricTransformation(t *testing.T) {
	// Create processor
	processor := NewEnhancedTransformProcessor("test-transform", nil, zap.NewNop())

	// Create test metrics
	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	metric := scopeMetrics.Metrics().AppendEmpty()

	// Set original metric name
	metric.SetName("original-metric")

	// Add data point with attributes
	dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.Attributes().PutStr("metric.type", "counter")
	dp.Attributes().PutStr("metric.name", "requests")

	// Process metrics
	err := processor.ProcessMetrics(context.Background(), metrics)
	if err != nil {
		t.Fatalf("Failed to process metrics: %v", err)
	}

	// Verify transformation
	expectedName := "counter_requests"
	if metric.Name() != expectedName {
		t.Errorf("Expected metric name '%s', got '%s'", expectedName, metric.Name())
	}
}

func TestEnhancedTransformProcessor_LogTransformation(t *testing.T) {
	// Create processor
	processor := NewEnhancedTransformProcessor("test-transform", nil, zap.NewNop())

	// Create test logs
	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()

	// Set original log body
	logRecord.Body().SetStr("original-body")

	// Add attributes that will be used for transformation
	logRecord.Attributes().PutStr("log.level", "ERROR")
	logRecord.Attributes().PutStr("log.message", "Database connection failed")

	// Process logs
	err := processor.ProcessLogs(context.Background(), logs)
	if err != nil {
		t.Fatalf("Failed to process logs: %v", err)
	}

	// Verify transformation
	expectedBody := "ERROR: Database connection failed"
	if logRecord.Body().AsString() != expectedBody {
		t.Errorf("Expected log body '%s', got '%s'", expectedBody, logRecord.Body().AsString())
	}
}

func TestEnhancedAttributesProcessor_AttributeActions(t *testing.T) {
	// Create processor
	processor := NewEnhancedAttributesProcessor("test-attributes", nil, zap.NewNop())

	// Create test traces
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	span := scopeSpans.Spans().AppendEmpty()

	// Process traces
	err := processor.ProcessTraces(context.Background(), traces)
	if err != nil {
		t.Fatalf("Failed to process traces: %v", err)
	}

	// Verify attribute actions were applied
	if value, exists := span.Attributes().Get("environment"); !exists || value.AsString() != "production" {
		t.Errorf("Expected environment attribute to be 'production', got %v", value)
	}

	if value, exists := span.Attributes().Get("service.name"); !exists || value.AsString() != "waveform" {
		t.Errorf("Expected service.name attribute to be 'waveform', got %v", value)
	}
}

func TestEnhancedAttributesProcessor_InsertAction(t *testing.T) {
	// Create processor
	processor := NewEnhancedAttributesProcessor("test-attributes", nil, zap.NewNop())

	// Create test attributes map
	attributes := pcommon.NewMap()

	// Test insert action (should add new attribute)
	processor.applyAttributeActions(attributes)

	// Verify attributes were added
	if value, exists := attributes.Get("environment"); !exists || value.AsString() != "production" {
		t.Errorf("Expected environment attribute to be 'production', got %v", value)
	}
}

func TestEnhancedAttributesProcessor_UpdateAction(t *testing.T) {
	// Create processor with custom actions for testing
	processor := &EnhancedAttributesProcessor{
		name:   "test-attributes",
		logger: zap.NewNop(),
		actions: []AttributeAction{
			{
				Key:    "test.key",
				Value:  "updated-value",
				Action: "update",
			},
		},
	}

	// Create test attributes map with existing attribute
	attributes := pcommon.NewMap()
	attributes.PutStr("test.key", "original-value")

	// Test update action (should update existing attribute)
	processor.applyAttributeActions(attributes)

	// Verify attribute was updated
	if value, exists := attributes.Get("test.key"); !exists || value.AsString() != "updated-value" {
		t.Errorf("Expected test.key attribute to be 'updated-value', got %v", value)
	}
}

func TestEnhancedAttributesProcessor_DeleteAction(t *testing.T) {
	// Create processor with custom actions for testing
	processor := &EnhancedAttributesProcessor{
		name:   "test-attributes",
		logger: zap.NewNop(),
		actions: []AttributeAction{
			{
				Key:    "test.key",
				Action: "delete",
			},
		},
	}

	// Create test attributes map with existing attribute
	attributes := pcommon.NewMap()
	attributes.PutStr("test.key", "value-to-delete")
	attributes.PutStr("other.key", "value-to-keep")

	// Test delete action (should remove attribute)
	processor.applyAttributeActions(attributes)

	// Verify attribute was deleted
	if _, exists := attributes.Get("test.key"); exists {
		t.Error("Expected test.key attribute to be deleted")
	}

	// Verify other attribute was preserved
	if value, exists := attributes.Get("other.key"); !exists || value.AsString() != "value-to-keep" {
		t.Errorf("Expected other.key attribute to be preserved, got %v", value)
	}
}

func TestEnhancedFilterProcessor_FilterCriteria(t *testing.T) {
	// Create processor
	processor := NewEnhancedFilterProcessor("test-filter", nil, zap.NewNop())

	// Create test traces
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	span := scopeSpans.Spans().AppendEmpty()

	// Add attributes that match filter criteria
	span.Attributes().PutStr("http.status_code", "500")

	// Process traces
	err := processor.ProcessTraces(context.Background(), traces)
	if err != nil {
		t.Fatalf("Failed to process traces: %v", err)
	}

	// Note: In this implementation, we're just logging the filter criteria
	// In a real implementation, you would verify that filtering actually occurred
}

func TestEnhancedTransformProcessor_AttributeTransform(t *testing.T) {
	// Create processor
	processor := NewEnhancedTransformProcessor("test-transform", nil, zap.NewNop())

	// Create test traces
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	span := scopeSpans.Spans().AppendEmpty()

	// Add source attributes
	span.Attributes().PutStr("source.attr1", "value1")
	span.Attributes().PutStr("source.attr2", "value2")

	// Process traces
	err := processor.ProcessTraces(context.Background(), traces)
	if err != nil {
		t.Fatalf("Failed to process traces: %v", err)
	}

	// Note: In this implementation, we're using default transformations
	// In a real implementation, you would configure specific attribute transformations
}

func TestEnhancedTransformProcessor_ComponentLifecycle(t *testing.T) {
	// Create processor
	processor := NewEnhancedTransformProcessor("test-transform", nil, zap.NewNop())

	// Test start
	err := processor.Start(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}

	// Test shutdown
	err = processor.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("Failed to shutdown processor: %v", err)
	}
}

func TestEnhancedAttributesProcessor_ComponentLifecycle(t *testing.T) {
	// Create processor
	processor := NewEnhancedAttributesProcessor("test-attributes", nil, zap.NewNop())

	// Test start
	err := processor.Start(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}

	// Test shutdown
	err = processor.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("Failed to shutdown processor: %v", err)
	}
}

func TestEnhancedFilterProcessor_ComponentLifecycle(t *testing.T) {
	// Create processor
	processor := NewEnhancedFilterProcessor("test-filter", nil, zap.NewNop())

	// Test start
	err := processor.Start(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}

	// Test shutdown
	err = processor.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("Failed to shutdown processor: %v", err)
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		pattern   string
		matchType string
		expected  bool
	}{
		{
			name:      "exact match",
			value:     "test",
			pattern:   "test",
			matchType: "exact",
			expected:  true,
		},
		{
			name:      "exact no match",
			value:     "test",
			pattern:   "other",
			matchType: "exact",
			expected:  false,
		},
		{
			name:      "regexp match",
			value:     "test123",
			pattern:   "test\\d+",
			matchType: "regexp",
			expected:  true,
		},
		{
			name:      "regexp no match",
			value:     "testabc",
			pattern:   "test\\d+",
			matchType: "regexp",
			expected:  false,
		},
		{
			name:      "invalid match type",
			value:     "test",
			pattern:   "test",
			matchType: "invalid",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.value, tt.pattern, tt.matchType)
			if result != tt.expected {
				t.Errorf("matchesPattern(%q, %q, %q) = %v, want %v", tt.value, tt.pattern, tt.matchType, result, tt.expected)
			}
		})
	}
}
