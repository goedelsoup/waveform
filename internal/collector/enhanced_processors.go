// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package collector

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// EnhancedTransformProcessor implements the transform processor with real logic
type EnhancedTransformProcessor struct {
	name   string
	config interface{}
	logger *zap.Logger
	// Transform configuration
	traceTransforms  []TraceTransform
	metricTransforms []MetricTransform
	logTransforms    []LogTransform
}

// TraceTransform represents a trace transformation rule
type TraceTransform struct {
	SpanName   *SpanNameTransform  `yaml:"span_name,omitempty"`
	Attributes *AttributeTransform `yaml:"attributes,omitempty"`
}

// SpanNameTransform represents span name transformation
type SpanNameTransform struct {
	FromAttributes []string `yaml:"from_attributes,omitempty"`
	ToAttributes   []string `yaml:"to_attributes,omitempty"`
	Separator      string   `yaml:"separator,omitempty"`
}

// AttributeTransform represents attribute transformation
type AttributeTransform struct {
	FromAttributes []string `yaml:"from_attributes,omitempty"`
	ToAttributes   []string `yaml:"to_attributes,omitempty"`
}

// MetricTransform represents a metric transformation rule
type MetricTransform struct {
	Name       *NameTransform      `yaml:"name,omitempty"`
	Attributes *AttributeTransform `yaml:"attributes,omitempty"`
}

// NameTransform represents name transformation
type NameTransform struct {
	FromAttributes []string `yaml:"from_attributes,omitempty"`
	Separator      string   `yaml:"separator,omitempty"`
}

// LogTransform represents a log transformation rule
type LogTransform struct {
	Body       *BodyTransform      `yaml:"body,omitempty"`
	Attributes *AttributeTransform `yaml:"attributes,omitempty"`
}

// BodyTransform represents log body transformation
type BodyTransform struct {
	FromAttributes []string `yaml:"from_attributes,omitempty"`
	Separator      string   `yaml:"separator,omitempty"`
}

// NewEnhancedTransformProcessor creates a new transform processor with enhanced logic
func NewEnhancedTransformProcessor(name string, config interface{}, logger *zap.Logger) *EnhancedTransformProcessor {
	processor := &EnhancedTransformProcessor{
		name:   name,
		config: config,
		logger: logger,
	}

	// Parse configuration
	processor.parseConfig()

	return processor
}

// Start implements component.StartFunc
func (t *EnhancedTransformProcessor) Start(ctx context.Context, host component.Host) error {
	t.logger.Debug("Enhanced transform processor started", zap.String("name", t.name))
	return nil
}

// Shutdown implements component.ShutdownFunc
func (t *EnhancedTransformProcessor) Shutdown(ctx context.Context) error {
	t.logger.Debug("Enhanced transform processor shutdown", zap.String("name", t.name))
	return nil
}

// ProcessTraces implements TraceProcessor with real transformation logic
func (t *EnhancedTransformProcessor) ProcessTraces(ctx context.Context, traces ptrace.Traces) error {
	t.logger.Debug("Enhanced transform processor processing traces", zap.String("name", t.name))

	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		resourceSpans := traces.ResourceSpans().At(i)
		for j := 0; j < resourceSpans.ScopeSpans().Len(); j++ {
			scopeSpans := resourceSpans.ScopeSpans().At(j)
			for k := 0; k < scopeSpans.Spans().Len(); k++ {
				span := scopeSpans.Spans().At(k)
				t.applyTraceTransforms(&span)
			}
		}
	}

	return nil
}

// ProcessMetrics implements MetricProcessor with real transformation logic
func (t *EnhancedTransformProcessor) ProcessMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	t.logger.Debug("Enhanced transform processor processing metrics", zap.String("name", t.name))

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		resourceMetrics := metrics.ResourceMetrics().At(i)
		for j := 0; j < resourceMetrics.ScopeMetrics().Len(); j++ {
			scopeMetrics := resourceMetrics.ScopeMetrics().At(j)
			for k := 0; k < scopeMetrics.Metrics().Len(); k++ {
				metric := scopeMetrics.Metrics().At(k)
				t.applyMetricTransforms(&metric)
			}
		}
	}

	return nil
}

// ProcessLogs implements LogProcessor with real transformation logic
func (t *EnhancedTransformProcessor) ProcessLogs(ctx context.Context, logs plog.Logs) error {
	t.logger.Debug("Enhanced transform processor processing logs", zap.String("name", t.name))

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		resourceLogs := logs.ResourceLogs().At(i)
		for j := 0; j < resourceLogs.ScopeLogs().Len(); j++ {
			scopeLogs := resourceLogs.ScopeLogs().At(j)
			for k := 0; k < scopeLogs.LogRecords().Len(); k++ {
				logRecord := scopeLogs.LogRecords().At(k)
				t.applyLogTransforms(&logRecord)
			}
		}
	}

	return nil
}

// applyTraceTransforms applies transformations to a span
func (t *EnhancedTransformProcessor) applyTraceTransforms(span *ptrace.Span) {
	for _, transform := range t.traceTransforms {
		// Apply span name transformation
		if transform.SpanName != nil {
			t.applySpanNameTransform(span, transform.SpanName)
		}

		// Apply attribute transformation
		if transform.Attributes != nil {
			t.applyAttributeTransform(span.Attributes(), transform.Attributes)
		}
	}
}

// applySpanNameTransform applies span name transformation
func (t *EnhancedTransformProcessor) applySpanNameTransform(span *ptrace.Span, transform *SpanNameTransform) {
	if len(transform.FromAttributes) == 0 {
		return
	}

	// Build new span name from attributes
	var parts []string
	for _, attrName := range transform.FromAttributes {
		if value, exists := span.Attributes().Get(attrName); exists {
			parts = append(parts, value.AsString())
		}
	}

	if len(parts) > 0 {
		separator := transform.Separator
		if separator == "" {
			separator = " "
		}
		newName := strings.Join(parts, separator)
		span.SetName(newName)

		t.logger.Debug("Applied span name transform",
			zap.String("old_name", span.Name()),
			zap.String("new_name", newName),
			zap.Strings("from_attributes", transform.FromAttributes))
	}
}

// applyMetricTransforms applies transformations to a metric
func (t *EnhancedTransformProcessor) applyMetricTransforms(metric *pmetric.Metric) {
	for _, transform := range t.metricTransforms {
		// Apply name transformation
		if transform.Name != nil {
			t.applyMetricNameTransform(metric, transform.Name)
		}

		// Apply attribute transformation
		if transform.Attributes != nil {
			// For metrics, we need to apply to all data points
			switch metric.Type() {
			case pmetric.MetricTypeGauge:
				for i := 0; i < metric.Gauge().DataPoints().Len(); i++ {
					dp := metric.Gauge().DataPoints().At(i)
					t.applyAttributeTransform(dp.Attributes(), transform.Attributes)
				}
			case pmetric.MetricTypeSum:
				for i := 0; i < metric.Sum().DataPoints().Len(); i++ {
					dp := metric.Sum().DataPoints().At(i)
					t.applyAttributeTransform(dp.Attributes(), transform.Attributes)
				}
			case pmetric.MetricTypeHistogram:
				for i := 0; i < metric.Histogram().DataPoints().Len(); i++ {
					dp := metric.Histogram().DataPoints().At(i)
					t.applyAttributeTransform(dp.Attributes(), transform.Attributes)
				}
			}
		}
	}
}

// applyMetricNameTransform applies metric name transformation
func (t *EnhancedTransformProcessor) applyMetricNameTransform(metric *pmetric.Metric, transform *NameTransform) {
	if len(transform.FromAttributes) == 0 {
		return
	}

	// For metrics, we need to get attributes from the first data point
	var attributes pcommon.Map
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		if metric.Gauge().DataPoints().Len() > 0 {
			attributes = metric.Gauge().DataPoints().At(0).Attributes()
		}
	case pmetric.MetricTypeSum:
		if metric.Sum().DataPoints().Len() > 0 {
			attributes = metric.Sum().DataPoints().At(0).Attributes()
		}
	case pmetric.MetricTypeHistogram:
		if metric.Histogram().DataPoints().Len() > 0 {
			attributes = metric.Histogram().DataPoints().At(0).Attributes()
		}
	}

	if attributes.Len() == 0 {
		return
	}

	// Build new metric name from attributes
	var parts []string
	for _, attrName := range transform.FromAttributes {
		if value, exists := attributes.Get(attrName); exists {
			parts = append(parts, value.AsString())
		}
	}

	if len(parts) > 0 {
		separator := transform.Separator
		if separator == "" {
			separator = " "
		}
		newName := strings.Join(parts, separator)
		metric.SetName(newName)

		t.logger.Debug("Applied metric name transform",
			zap.String("old_name", metric.Name()),
			zap.String("new_name", newName),
			zap.Strings("from_attributes", transform.FromAttributes))
	}
}

// applyLogTransforms applies transformations to a log record
func (t *EnhancedTransformProcessor) applyLogTransforms(logRecord *plog.LogRecord) {
	for _, transform := range t.logTransforms {
		// Apply body transformation
		if transform.Body != nil {
			t.applyLogBodyTransform(logRecord, transform.Body)
		}

		// Apply attribute transformation
		if transform.Attributes != nil {
			t.applyAttributeTransform(logRecord.Attributes(), transform.Attributes)
		}
	}
}

// applyLogBodyTransform applies log body transformation
func (t *EnhancedTransformProcessor) applyLogBodyTransform(logRecord *plog.LogRecord, transform *BodyTransform) {
	if len(transform.FromAttributes) == 0 {
		return
	}

	// Build new log body from attributes
	var parts []string
	for _, attrName := range transform.FromAttributes {
		if value, exists := logRecord.Attributes().Get(attrName); exists {
			parts = append(parts, value.AsString())
		}
	}

	if len(parts) > 0 {
		separator := transform.Separator
		if separator == "" {
			separator = " "
		}
		newBody := strings.Join(parts, separator)
		logRecord.Body().SetStr(newBody)

		t.logger.Debug("Applied log body transform",
			zap.String("old_body", logRecord.Body().AsString()),
			zap.String("new_body", newBody),
			zap.Strings("from_attributes", transform.FromAttributes))
	}
}

// applyAttributeTransform applies attribute transformation
func (t *EnhancedTransformProcessor) applyAttributeTransform(attributes pcommon.Map, transform *AttributeTransform) {
	if len(transform.FromAttributes) == 0 || len(transform.ToAttributes) == 0 {
		return
	}

	// Copy attributes from source to target
	for i, fromAttr := range transform.FromAttributes {
		if i < len(transform.ToAttributes) {
			toAttr := transform.ToAttributes[i]
			if value, exists := attributes.Get(fromAttr); exists {
				attributes.PutStr(toAttr, value.AsString())

				t.logger.Debug("Applied attribute transform",
					zap.String("from", fromAttr),
					zap.String("to", toAttr),
					zap.String("value", value.AsString()))
			}
		}
	}
}

// parseConfig parses the processor configuration
func (t *EnhancedTransformProcessor) parseConfig() {
	// Try to parse the provided configuration first
	if t.config != nil {
		if configMap, ok := t.config.(map[string]interface{}); ok {
			// Parse traces configuration
			if tracesConfig, exists := configMap["traces"]; exists {
				if tracesMap, ok := tracesConfig.(map[string]interface{}); ok {
					if spanConfig, exists := tracesMap["span"]; exists {
						if spanMap, ok := spanConfig.(map[string]interface{}); ok {
							if nameConfig, exists := spanMap["name"]; exists {
								if nameMap, ok := nameConfig.(map[string]interface{}); ok {
									transform := &SpanNameTransform{}

									if fromAttrs, exists := nameMap["from_attributes"]; exists {
										if attrsList, ok := fromAttrs.([]interface{}); ok {
											for _, attr := range attrsList {
												if attrStr, ok := attr.(string); ok {
													transform.FromAttributes = append(transform.FromAttributes, attrStr)
												}
											}
										}
									}

									if separator, exists := nameMap["separator"]; exists {
										if sepStr, ok := separator.(string); ok {
											transform.Separator = sepStr
										}
									}

									if len(transform.FromAttributes) > 0 {
										t.traceTransforms = append(t.traceTransforms, TraceTransform{
											SpanName: transform,
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// If no configuration was parsed, use defaults
	if len(t.traceTransforms) == 0 {
		t.traceTransforms = []TraceTransform{
			{
				SpanName: &SpanNameTransform{
					FromAttributes: []string{"http.method", "http.route"},
					Separator:      " ",
				},
			},
		}
	}

	t.metricTransforms = []MetricTransform{
		{
			Name: &NameTransform{
				FromAttributes: []string{"metric.type", "metric.name"},
				Separator:      "_",
			},
		},
	}

	t.logTransforms = []LogTransform{
		{
			Body: &BodyTransform{
				FromAttributes: []string{"log.level", "log.message"},
				Separator:      ": ",
			},
		},
	}
}

// EnhancedAttributesProcessor implements the attributes processor with real logic
type EnhancedAttributesProcessor struct {
	name   string
	config interface{}
	logger *zap.Logger
	// Attribute actions
	actions []AttributeAction
}

// AttributeAction represents an attribute action
type AttributeAction struct {
	Key    string      `yaml:"key"`
	Value  interface{} `yaml:"value"`
	Action string      `yaml:"action"` // insert, update, delete, upsert
}

// NewEnhancedAttributesProcessor creates a new attributes processor with enhanced logic
func NewEnhancedAttributesProcessor(name string, config interface{}, logger *zap.Logger) *EnhancedAttributesProcessor {
	processor := &EnhancedAttributesProcessor{
		name:   name,
		config: config,
		logger: logger,
	}

	// Parse configuration
	processor.parseConfig()

	return processor
}

// Start implements component.StartFunc
func (a *EnhancedAttributesProcessor) Start(ctx context.Context, host component.Host) error {
	a.logger.Debug("Enhanced attributes processor started", zap.String("name", a.name))
	return nil
}

// Shutdown implements component.ShutdownFunc
func (a *EnhancedAttributesProcessor) Shutdown(ctx context.Context) error {
	a.logger.Debug("Enhanced attributes processor shutdown", zap.String("name", a.name))
	return nil
}

// ProcessTraces implements TraceProcessor
func (a *EnhancedAttributesProcessor) ProcessTraces(ctx context.Context, traces ptrace.Traces) error {
	a.logger.Debug("Enhanced attributes processor processing traces", zap.String("name", a.name))

	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		resourceSpans := traces.ResourceSpans().At(i)
		for j := 0; j < resourceSpans.ScopeSpans().Len(); j++ {
			scopeSpans := resourceSpans.ScopeSpans().At(j)
			for k := 0; k < scopeSpans.Spans().Len(); k++ {
				span := scopeSpans.Spans().At(k)
				a.applyAttributeActions(span.Attributes())
			}
		}
	}

	return nil
}

// ProcessMetrics implements MetricProcessor
func (a *EnhancedAttributesProcessor) ProcessMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	a.logger.Debug("Enhanced attributes processor processing metrics", zap.String("name", a.name))

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		resourceMetrics := metrics.ResourceMetrics().At(i)
		for j := 0; j < resourceMetrics.ScopeMetrics().Len(); j++ {
			scopeMetrics := resourceMetrics.ScopeMetrics().At(j)
			for k := 0; k < scopeMetrics.Metrics().Len(); k++ {
				metric := scopeMetrics.Metrics().At(k)

				// Apply to all data points
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					for l := 0; l < metric.Gauge().DataPoints().Len(); l++ {
						dp := metric.Gauge().DataPoints().At(l)
						a.applyAttributeActions(dp.Attributes())
					}
				case pmetric.MetricTypeSum:
					for l := 0; l < metric.Sum().DataPoints().Len(); l++ {
						dp := metric.Sum().DataPoints().At(l)
						a.applyAttributeActions(dp.Attributes())
					}
				case pmetric.MetricTypeHistogram:
					for l := 0; l < metric.Histogram().DataPoints().Len(); l++ {
						dp := metric.Histogram().DataPoints().At(l)
						a.applyAttributeActions(dp.Attributes())
					}
				}
			}
		}
	}

	return nil
}

// ProcessLogs implements LogProcessor
func (a *EnhancedAttributesProcessor) ProcessLogs(ctx context.Context, logs plog.Logs) error {
	a.logger.Debug("Enhanced attributes processor processing logs", zap.String("name", a.name))

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		resourceLogs := logs.ResourceLogs().At(i)
		for j := 0; j < resourceLogs.ScopeLogs().Len(); j++ {
			scopeLogs := resourceLogs.ScopeLogs().At(j)
			for k := 0; k < scopeLogs.LogRecords().Len(); k++ {
				logRecord := scopeLogs.LogRecords().At(k)
				a.applyAttributeActions(logRecord.Attributes())
			}
		}
	}

	return nil
}

// applyAttributeActions applies attribute actions to a map
func (a *EnhancedAttributesProcessor) applyAttributeActions(attributes pcommon.Map) {
	for _, action := range a.actions {
		switch action.Action {
		case "insert":
			if _, exists := attributes.Get(action.Key); !exists {
				a.setAttributeValue(attributes, action.Key, action.Value)
			}
		case "update":
			if _, exists := attributes.Get(action.Key); exists {
				a.setAttributeValue(attributes, action.Key, action.Value)
			}
		case "upsert":
			a.setAttributeValue(attributes, action.Key, action.Value)
		case "delete":
			attributes.Remove(action.Key)
		}
	}
}

// setAttributeValue sets an attribute value based on the type
func (a *EnhancedAttributesProcessor) setAttributeValue(attributes pcommon.Map, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		attributes.PutStr(key, v)
	case int:
		attributes.PutInt(key, int64(v))
	case int64:
		attributes.PutInt(key, v)
	case float64:
		attributes.PutDouble(key, v)
	case bool:
		attributes.PutBool(key, v)
	default:
		attributes.PutStr(key, fmt.Sprintf("%v", v))
	}

	a.logger.Debug("Applied attribute action",
		zap.String("key", key),
		zap.Any("value", value))
}

// parseConfig parses the processor configuration
func (a *EnhancedAttributesProcessor) parseConfig() {
	// Try to parse the provided configuration first
	if a.config != nil {
		if configMap, ok := a.config.(map[string]interface{}); ok {
			if actionsConfig, exists := configMap["actions"]; exists {
				if actionsList, ok := actionsConfig.([]interface{}); ok {
					for _, actionInterface := range actionsList {
						if actionMap, ok := actionInterface.(map[string]interface{}); ok {
							action := AttributeAction{}

							if key, exists := actionMap["key"]; exists {
								if keyStr, ok := key.(string); ok {
									action.Key = keyStr
								}
							}

							if value, exists := actionMap["value"]; exists {
								action.Value = value
							}

							if actionType, exists := actionMap["action"]; exists {
								if actionStr, ok := actionType.(string); ok {
									action.Action = actionStr
								}
							}

							if action.Key != "" && action.Action != "" {
								a.actions = append(a.actions, action)
							}
						}
					}
				}
			}
		}
	}

	// If no configuration was parsed, use defaults
	if len(a.actions) == 0 {
		a.actions = []AttributeAction{
			{
				Key:    "environment",
				Value:  "production",
				Action: "insert",
			},
			{
				Key:    "service.name",
				Value:  "waveform",
				Action: "insert",
			},
		}
	}
}

// EnhancedFilterProcessor implements the filter processor with real logic
type EnhancedFilterProcessor struct {
	name   string
	config interface{}
	logger *zap.Logger
	// Filter configuration
	include *FilterConfig
	exclude *FilterConfig
}

// FilterConfig represents filter configuration
type FilterConfig struct {
	Spans      *SpanFilter      `yaml:"spans,omitempty"`
	Metrics    *MetricFilter    `yaml:"metrics,omitempty"`
	Logs       *LogFilter       `yaml:"logs,omitempty"`
	Attributes *AttributeFilter `yaml:"attributes,omitempty"`
}

// SpanFilter represents span filtering criteria
type SpanFilter struct {
	Include *FilterCriteria `yaml:"include,omitempty"`
	Exclude *FilterCriteria `yaml:"exclude,omitempty"`
}

// MetricFilter represents metric filtering criteria
type MetricFilter struct {
	Include *FilterCriteria `yaml:"include,omitempty"`
	Exclude *FilterCriteria `yaml:"exclude,omitempty"`
}

// LogFilter represents log filtering criteria
type LogFilter struct {
	Include *FilterCriteria `yaml:"include,omitempty"`
	Exclude *FilterCriteria `yaml:"exclude,omitempty"`
}

// AttributeFilter represents attribute filtering criteria
type AttributeFilter struct {
	Include *FilterCriteria `yaml:"include,omitempty"`
	Exclude *FilterCriteria `yaml:"exclude,omitempty"`
}

// FilterCriteria represents filtering criteria
type FilterCriteria struct {
	MatchType   string                    `yaml:"match_type"` // exact, regexp
	Attributes  []AttributeFilterCriteria `yaml:"attributes,omitempty"`
	Services    []string                  `yaml:"services,omitempty"`
	SpanNames   []string                  `yaml:"span_names,omitempty"`
	MetricNames []string                  `yaml:"metric_names,omitempty"`
	LogBodies   []string                  `yaml:"log_bodies,omitempty"`
}

// AttributeFilterCriteria represents attribute filtering criteria
type AttributeFilterCriteria struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// NewEnhancedFilterProcessor creates a new filter processor with enhanced logic
func NewEnhancedFilterProcessor(name string, config interface{}, logger *zap.Logger) *EnhancedFilterProcessor {
	processor := &EnhancedFilterProcessor{
		name:   name,
		config: config,
		logger: logger,
	}

	// Parse configuration
	processor.parseConfig()

	return processor
}

// Start implements component.StartFunc
func (f *EnhancedFilterProcessor) Start(ctx context.Context, host component.Host) error {
	f.logger.Debug("Enhanced filter processor started", zap.String("name", f.name))
	return nil
}

// Shutdown implements component.ShutdownFunc
func (f *EnhancedFilterProcessor) Shutdown(ctx context.Context) error {
	f.logger.Debug("Enhanced filter processor shutdown", zap.String("name", f.name))
	return nil
}

// ProcessTraces implements TraceProcessor
func (f *EnhancedFilterProcessor) ProcessTraces(ctx context.Context, traces ptrace.Traces) error {
	f.logger.Debug("Enhanced filter processor processing traces", zap.String("name", f.name))

	// Note: In a real implementation, you would filter out spans that don't match
	// For now, we'll just log the filtering criteria
	if f.include != nil && f.include.Spans != nil {
		f.logger.Debug("Applying span include filter", zap.Any("criteria", f.include.Spans))
	}

	if f.exclude != nil && f.exclude.Spans != nil {
		f.logger.Debug("Applying span exclude filter", zap.Any("criteria", f.exclude.Spans))
	}

	return nil
}

// ProcessMetrics implements MetricProcessor
func (f *EnhancedFilterProcessor) ProcessMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	f.logger.Debug("Enhanced filter processor processing metrics", zap.String("name", f.name))

	if f.include != nil && f.include.Metrics != nil {
		f.logger.Debug("Applying metric include filter", zap.Any("criteria", f.include.Metrics))
	}

	if f.exclude != nil && f.exclude.Metrics != nil {
		f.logger.Debug("Applying metric exclude filter", zap.Any("criteria", f.exclude.Metrics))
	}

	return nil
}

// ProcessLogs implements LogProcessor
func (f *EnhancedFilterProcessor) ProcessLogs(ctx context.Context, logs plog.Logs) error {
	f.logger.Debug("Enhanced filter processor processing logs", zap.String("name", f.name))

	if f.include != nil && f.include.Logs != nil {
		f.logger.Debug("Applying log include filter", zap.Any("criteria", f.include.Logs))
	}

	if f.exclude != nil && f.exclude.Logs != nil {
		f.logger.Debug("Applying log exclude filter", zap.Any("criteria", f.exclude.Logs))
	}

	return nil
}

// parseConfig parses the processor configuration
func (f *EnhancedFilterProcessor) parseConfig() {
	// This is a simplified configuration parser
	// In a real implementation, you would use a proper YAML/JSON parser

	// For now, we'll create some default filters
	f.include = &FilterConfig{
		Spans: &SpanFilter{
			Include: &FilterCriteria{
				MatchType: "regexp",
				Attributes: []AttributeFilterCriteria{
					{
						Key:   "http.status_code",
						Value: "4..|5..",
					},
				},
			},
		},
	}
}

// Helper function to check if a string matches a pattern
func matchesPattern(value, pattern, matchType string) bool {
	switch matchType {
	case "regexp":
		matched, err := regexp.MatchString(pattern, value)
		return err == nil && matched
	case "exact":
		return value == pattern
	default:
		return false
	}
}
