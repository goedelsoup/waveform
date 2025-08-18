// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package generator

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goedelsoup/waveform/internal/contract"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// Generator creates realistic OpenTelemetry data based on contracts
type Generator struct {
	baseTime time.Time
}

// NewGenerator creates a new mock data generator
func NewGenerator() *Generator {
	return &Generator{
		baseTime: time.Now(),
	}
}

// SetBaseTime sets the base time for generating timestamps
func (g *Generator) SetBaseTime(baseTime time.Time) {
	g.baseTime = baseTime
}

// GenerateFromContract generates OpenTelemetry data based on a contract
func (g *Generator) GenerateFromContract(contractDef *contract.Contract) contract.OpenTelemetryData {
	var data contract.OpenTelemetryData
	data.Time = g.baseTime

	// Generate traces
	if len(contractDef.Inputs.Traces) > 0 {
		data.Traces = g.generateTraces(contractDef.Inputs.Traces)
	}

	// Generate metrics
	if len(contractDef.Inputs.Metrics) > 0 {
		data.Metrics = g.generateMetrics(contractDef.Inputs.Metrics)
	}

	// Generate logs
	if len(contractDef.Inputs.Logs) > 0 {
		data.Logs = g.generateLogs(contractDef.Inputs.Logs)
	}

	return data
}

// GenerateRealistic generates realistic OpenTelemetry data for a signal type
func (g *Generator) GenerateRealistic(signalType contract.SignalType, count int) contract.OpenTelemetryData {
	data := contract.OpenTelemetryData{
		Time: g.baseTime,
	}

	switch signalType {
	case contract.SignalTypeTraces:
		data.Traces = g.generateRealisticTraces(count)
	case contract.SignalTypeMetrics:
		data.Metrics = g.generateRealisticMetrics(count)
	case contract.SignalTypeLogs:
		data.Logs = g.generateRealisticLogs(count)
	}

	return data
}

// generateTraces generates trace data from contract inputs
func (g *Generator) generateTraces(inputs []contract.TraceInput) ptrace.Traces {
	traces := ptrace.NewTraces()

	for _, input := range inputs {
		resourceSpans := traces.ResourceSpans().AppendEmpty()

		// Set resource attributes
		if input.ServiceName != "" {
			resourceSpans.Resource().Attributes().PutStr("service.name", input.ServiceName)
		}

		scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
		span := scopeSpans.Spans().AppendEmpty()

		// Set span name
		span.SetName(input.SpanName)

		// Set span kind (default to server)
		span.SetKind(ptrace.SpanKindServer)

		// Set timestamps
		span.SetStartTimestamp(pcommon.NewTimestampFromTime(g.baseTime))
		span.SetEndTimestamp(pcommon.NewTimestampFromTime(g.baseTime.Add(100 * time.Millisecond)))

		// Generate trace and span IDs
		traceID := g.generateTraceID()
		spanID := g.generateSpanID()
		span.SetTraceID(traceID)
		span.SetSpanID(spanID)

		// Set parent span if specified
		if input.ParentSpan != "" {
			parentSpanID := g.generateSpanID()
			span.SetParentSpanID(parentSpanID)
		}

		// Set attributes
		for key, value := range input.Attributes {
			g.setAttribute(span.Attributes(), key, value)
		}
	}

	return traces
}

// generateMetrics generates metric data from contract inputs
func (g *Generator) generateMetrics(inputs []contract.MetricInput) pmetric.Metrics {
	metrics := pmetric.NewMetrics()

	for _, input := range inputs {
		resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
		metric := scopeMetrics.Metrics().AppendEmpty()

		// Set metric name
		metric.SetName(input.Name)

		// Set metric type
		metricType := input.Type
		if metricType == "" {
			metricType = "counter"
		}

		var dataPoint pmetric.NumberDataPoint

		switch strings.ToLower(metricType) {
		case "counter":
			metric.SetEmptySum()
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			dataPoint = metric.Sum().DataPoints().AppendEmpty()
			g.setMetricValue(dataPoint, input.Value)
		case "gauge":
			metric.SetEmptyGauge()
			dataPoint = metric.Gauge().DataPoints().AppendEmpty()
			g.setMetricValue(dataPoint, input.Value)
		case "histogram":
			metric.SetEmptyHistogram()
			metric.Histogram().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			histogramDataPoint := metric.Histogram().DataPoints().AppendEmpty()
			// For histogram, we'll create a simple distribution
			if val, ok := input.Value.(float64); ok {
				histogramDataPoint.SetCount(1)
				histogramDataPoint.SetSum(val)
				histogramDataPoint.SetMin(val)
				histogramDataPoint.SetMax(val)
			}
			// Set labels/attributes for histogram
			for key, value := range input.Labels {
				g.setAttribute(histogramDataPoint.Attributes(), key, value)
			}
			// Set timestamp for histogram
			histogramDataPoint.SetTimestamp(pcommon.NewTimestampFromTime(g.baseTime))
			continue
		}

		// Set labels/attributes
		for key, value := range input.Labels {
			g.setAttribute(dataPoint.Attributes(), key, value)
		}

		// Set timestamp
		dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(g.baseTime))
	}

	return metrics
}

// generateLogs generates log data from contract inputs
func (g *Generator) generateLogs(inputs []contract.LogInput) plog.Logs {
	logs := plog.NewLogs()

	for _, input := range inputs {
		resourceLogs := logs.ResourceLogs().AppendEmpty()
		scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
		logRecord := scopeLogs.LogRecords().AppendEmpty()

		// Set log body
		logRecord.Body().SetStr(input.Body)

		// Set severity
		if input.Severity != "" {
			severity := g.parseSeverity(input.Severity)
			logRecord.SetSeverityNumber(severity)
			logRecord.SetSeverityText(input.Severity)
		}

		// Set attributes
		for key, value := range input.Attributes {
			g.setAttribute(logRecord.Attributes(), key, value)
		}

		// Set timestamp
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(g.baseTime))

		// Set trace context if available
		traceID := g.generateTraceID()
		spanID := g.generateSpanID()
		logRecord.SetTraceID(traceID)
		logRecord.SetSpanID(spanID)
	}

	return logs
}

// generateRealisticTraces generates realistic trace data
func (g *Generator) generateRealisticTraces(count int) ptrace.Traces {
	traces := ptrace.NewTraces()

	for i := 0; i < count; i++ {
		resourceSpans := traces.ResourceSpans().AppendEmpty()
		resourceSpans.Resource().Attributes().PutStr("service.name", fmt.Sprintf("service-%d", i))

		scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
		span := scopeSpans.Spans().AppendEmpty()

		span.SetName(fmt.Sprintf("operation-%d", i))
		span.SetKind(ptrace.SpanKindServer)
		span.SetStartTimestamp(pcommon.NewTimestampFromTime(g.baseTime.Add(time.Duration(i) * time.Millisecond)))
		span.SetEndTimestamp(pcommon.NewTimestampFromTime(g.baseTime.Add(time.Duration(i+1) * 100 * time.Millisecond)))

		traceID := g.generateTraceID()
		spanID := g.generateSpanID()
		span.SetTraceID(traceID)
		span.SetSpanID(spanID)

		// Add some realistic attributes
		span.Attributes().PutStr("http.method", "GET")
		span.Attributes().PutStr("http.url", fmt.Sprintf("/api/v1/resource/%d", i))
		span.Attributes().PutInt("http.status_code", 200)
	}

	return traces
}

// generateRealisticMetrics generates realistic metric data
func (g *Generator) generateRealisticMetrics(count int) pmetric.Metrics {
	metrics := pmetric.NewMetrics()

	for i := 0; i < count; i++ {
		resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
		scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
		metric := scopeMetrics.Metrics().AppendEmpty()

		metric.SetName(fmt.Sprintf("metric_%d", i))
		metric.SetEmptyGauge()

		dataPoint := metric.Gauge().DataPoints().AppendEmpty()
		dataPoint.SetDoubleValue(float64(i) + 1.0)
		dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(g.baseTime))

		// Add labels
		dataPoint.Attributes().PutStr("label1", fmt.Sprintf("value_%d", i))
		dataPoint.Attributes().PutStr("label2", "constant")
	}

	return metrics
}

// generateRealisticLogs generates realistic log data
func (g *Generator) generateRealisticLogs(count int) plog.Logs {
	logs := plog.NewLogs()

	for i := 0; i < count; i++ {
		resourceLogs := logs.ResourceLogs().AppendEmpty()
		scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
		logRecord := scopeLogs.LogRecords().AppendEmpty()

		logRecord.Body().SetStr(fmt.Sprintf("Log message %d", i))
		logRecord.SetSeverityNumber(plog.SeverityNumberInfo)
		logRecord.SetSeverityText("INFO")
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(g.baseTime.Add(time.Duration(i) * time.Millisecond)))

		// Add some realistic attributes
		logRecord.Attributes().PutStr("logger", fmt.Sprintf("logger_%d", i))
		logRecord.Attributes().PutStr("thread", "main")
	}

	return logs
}

// Helper methods

func (g *Generator) generateTraceID() pcommon.TraceID {
	var traceID [16]byte
	if _, err := rand.Read(traceID[:]); err != nil {
		// In a real implementation, you might want to handle this error more gracefully
		// For now, we'll use a fallback approach
		for i := range traceID {
			traceID[i] = byte(i)
		}
	}
	return pcommon.TraceID(traceID)
}

func (g *Generator) generateSpanID() pcommon.SpanID {
	var spanID [8]byte
	if _, err := rand.Read(spanID[:]); err != nil {
		// In a real implementation, you might want to handle this error more gracefully
		// For now, we'll use a fallback approach
		for i := range spanID {
			spanID[i] = byte(i)
		}
	}
	return pcommon.SpanID(spanID)
}

func (g *Generator) setAttribute(attrs pcommon.Map, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		attrs.PutStr(key, v)
	case int:
		attrs.PutInt(key, int64(v))
	case int64:
		attrs.PutInt(key, v)
	case float64:
		attrs.PutDouble(key, v)
	case bool:
		attrs.PutBool(key, v)
	case map[string]interface{}:
		nestedMap := attrs.PutEmptyMap(key)
		for nestedKey, nestedValue := range v {
			g.setAttribute(nestedMap, nestedKey, nestedValue)
		}
	case []interface{}:
		nestedSlice := attrs.PutEmptySlice(key)
		for _, item := range v {
			element := nestedSlice.AppendEmpty()
			g.setValue(element, item)
		}
	default:
		// Try to convert to string
		attrs.PutStr(key, fmt.Sprintf("%v", v))
	}
}

func (g *Generator) setValue(value pcommon.Value, item interface{}) {
	switch v := item.(type) {
	case string:
		value.SetStr(v)
	case int:
		value.SetInt(int64(v))
	case int64:
		value.SetInt(v)
	case float64:
		value.SetDouble(v)
	case bool:
		value.SetBool(v)
	case map[string]interface{}:
		nestedMap := value.SetEmptyMap()
		for nestedKey, nestedValue := range v {
			g.setAttribute(nestedMap, nestedKey, nestedValue)
		}
	case []interface{}:
		nestedSlice := value.SetEmptySlice()
		for _, sliceItem := range v {
			element := nestedSlice.AppendEmpty()
			g.setValue(element, sliceItem)
		}
	default:
		value.SetStr(fmt.Sprintf("%v", v))
	}
}

func (g *Generator) setMetricValue(dataPoint pmetric.NumberDataPoint, value interface{}) {
	switch v := value.(type) {
	case int:
		dataPoint.SetIntValue(int64(v))
	case int64:
		dataPoint.SetIntValue(v)
	case float64:
		dataPoint.SetDoubleValue(v)
	case string:
		// Try to parse as number
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			dataPoint.SetDoubleValue(f)
		} else {
			dataPoint.SetDoubleValue(0)
		}
	default:
		dataPoint.SetDoubleValue(0)
	}
}

func (g *Generator) parseSeverity(severity string) plog.SeverityNumber {
	switch strings.ToUpper(severity) {
	case "TRACE":
		return plog.SeverityNumberTrace
	case "DEBUG":
		return plog.SeverityNumberDebug
	case "INFO":
		return plog.SeverityNumberInfo
	case "WARN", "WARNING":
		return plog.SeverityNumberWarn
	case "ERROR":
		return plog.SeverityNumberError
	case "FATAL":
		return plog.SeverityNumberFatal
	default:
		return plog.SeverityNumberInfo
	}
}
