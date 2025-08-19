// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package collector

import (
	"context"
	"fmt"

	"github.com/goedelsoup/waveform/internal/contract"
	"github.com/goedelsoup/waveform/internal/harness"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// TraceProcessor defines the interface for trace processing
type TraceProcessor interface {
	ProcessTraces(ctx context.Context, traces ptrace.Traces) error
}

// MetricProcessor defines the interface for metric processing
type MetricProcessor interface {
	ProcessMetrics(ctx context.Context, metrics pmetric.Metrics) error
}

// LogProcessor defines the interface for log processing
type LogProcessor interface {
	ProcessLogs(ctx context.Context, logs plog.Logs) error
}

// Service represents a running OpenTelemetry collector service
type Service struct {
	config     harness.CollectorConfig
	logger     *zap.Logger
	receivers  map[string]component.Component
	processors map[string]component.Component
	exporters  map[string]component.Component
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewService creates a new collector service
func NewService(config harness.CollectorConfig, logger *zap.Logger) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		config:     config,
		logger:     logger,
		receivers:  make(map[string]component.Component),
		processors: make(map[string]component.Component),
		exporters:  make(map[string]component.Component),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start initializes and starts the collector service
func (s *Service) Start() error {
	s.logger.Info("Starting collector service")

	// Initialize components
	if err := s.initializeComponents(); err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Start all components
	if err := s.startComponents(); err != nil {
		return fmt.Errorf("failed to start components: %w", err)
	}

	s.logger.Info("Collector service started successfully")
	return nil
}

// Stop gracefully shuts down the collector service
func (s *Service) Stop() error {
	s.logger.Info("Stopping collector service")

	// Cancel context to signal shutdown
	s.cancel()

	// Stop all components
	if err := s.stopComponents(); err != nil {
		return fmt.Errorf("failed to stop components: %w", err)
	}

	s.logger.Info("Collector service stopped successfully")
	return nil
}

// ProcessData processes telemetry data through the configured pipeline
func (s *Service) ProcessData(input interface{}) (interface{}, error) {
	s.logger.Debug("Processing telemetry data through collector")

	// Convert input to OpenTelemetryData if possible
	var data contract.OpenTelemetryData
	if inputData, ok := input.(contract.OpenTelemetryData); ok {
		data = inputData
	} else {
		// If conversion fails, return input unchanged
		return input, nil
	}

	// Process traces through processors
	if data.Traces.ResourceSpans().Len() > 0 {
		for _, processor := range s.processors {
			if traceProcessor, ok := processor.(TraceProcessor); ok {
				if err := traceProcessor.ProcessTraces(s.ctx, data.Traces); err != nil {
					s.logger.Error("Failed to process traces", zap.Error(err))
				}
			}
		}
	}

	// Process metrics through processors
	if data.Metrics.ResourceMetrics().Len() > 0 {
		for _, processor := range s.processors {
			if metricProcessor, ok := processor.(MetricProcessor); ok {
				if err := metricProcessor.ProcessMetrics(s.ctx, data.Metrics); err != nil {
					s.logger.Error("Failed to process metrics", zap.Error(err))
				}
			}
		}
	}

	// Process logs through processors
	if data.Logs.ResourceLogs().Len() > 0 {
		for _, processor := range s.processors {
			if logProcessor, ok := processor.(LogProcessor); ok {
				if err := logProcessor.ProcessLogs(s.ctx, data.Logs); err != nil {
					s.logger.Error("Failed to process logs", zap.Error(err))
				}
			}
		}
	}

	return data, nil
}

// initializeComponents initializes all collector components
func (s *Service) initializeComponents() error {
	s.logger.Debug("Initializing collector components")

	// Initialize receivers
	for name, config := range s.config.Receivers {
		if err := s.initializeReceiver(name, config); err != nil {
			return fmt.Errorf("failed to initialize receiver %s: %w", name, err)
		}
	}

	// Initialize processors
	for name, config := range s.config.Processors {
		if err := s.initializeProcessor(name, config); err != nil {
			return fmt.Errorf("failed to initialize processor %s: %w", name, err)
		}
	}

	// Initialize exporters
	for name, config := range s.config.Exporters {
		if err := s.initializeExporter(name, config); err != nil {
			return fmt.Errorf("failed to initialize exporter %s: %w", name, err)
		}
	}

	return nil
}

// startComponents starts all collector components
func (s *Service) startComponents() error {
	s.logger.Debug("Starting collector components")

	// Start receivers
	for name, receiver := range s.receivers {
		if starter, ok := receiver.(interface {
			Start(context.Context, component.Host) error
		}); ok {
			if err := starter.Start(s.ctx, nil); err != nil {
				return fmt.Errorf("failed to start receiver %s: %w", name, err)
			}
		}
	}

	// Start processors
	for name, processor := range s.processors {
		if starter, ok := processor.(interface {
			Start(context.Context, component.Host) error
		}); ok {
			if err := starter.Start(s.ctx, nil); err != nil {
				return fmt.Errorf("failed to start processor %s: %w", name, err)
			}
		}
	}

	// Start exporters
	for name, exporter := range s.exporters {
		if starter, ok := exporter.(interface {
			Start(context.Context, component.Host) error
		}); ok {
			if err := starter.Start(s.ctx, nil); err != nil {
				return fmt.Errorf("failed to start exporter %s: %w", name, err)
			}
		}
	}

	return nil
}

// stopComponents stops all collector components
func (s *Service) stopComponents() error {
	s.logger.Debug("Stopping collector components")

	// Stop exporters
	for name, exporter := range s.exporters {
		if stopper, ok := exporter.(interface{ Shutdown(context.Context) error }); ok {
			if err := stopper.Shutdown(s.ctx); err != nil {
				return fmt.Errorf("failed to stop exporter %s: %w", name, err)
			}
		}
	}

	// Stop processors
	for name, processor := range s.processors {
		if stopper, ok := processor.(interface{ Shutdown(context.Context) error }); ok {
			if err := stopper.Shutdown(s.ctx); err != nil {
				return fmt.Errorf("failed to stop processor %s: %w", name, err)
			}
		}
	}

	// Stop receivers
	for name, receiver := range s.receivers {
		if stopper, ok := receiver.(interface{ Shutdown(context.Context) error }); ok {
			if err := stopper.Shutdown(s.ctx); err != nil {
				return fmt.Errorf("failed to stop receiver %s: %w", name, err)
			}
		}
	}

	return nil
}

// initializeReceiver initializes a receiver component
func (s *Service) initializeReceiver(name string, config interface{}) error {
	s.logger.Debug("Initializing receiver", zap.String("name", name))

	// For now, create a mock receiver
	// In a real implementation, this would create actual receiver instances
	receiver := &MockReceiver{
		name:   name,
		config: config,
		logger: s.logger,
	}

	s.receivers[name] = receiver
	return nil
}

// initializeProcessor initializes a processor component
func (s *Service) initializeProcessor(name string, config interface{}) error {
	s.logger.Debug("Initializing processor", zap.String("name", name))

	// Create processor based on type
	var processor component.Component
	switch name {
	case "transform":
		processor = NewEnhancedTransformProcessor(name, config, s.logger)
	case "attributes":
		processor = NewEnhancedAttributesProcessor(name, config, s.logger)
	case "filter":
		processor = NewEnhancedFilterProcessor(name, config, s.logger)
	case "batch":
		processor = NewBatchProcessor(name, config, s.logger)
	case "memory_limiter":
		processor = NewMemoryLimiterProcessor(name, config, s.logger)
	case "resource":
		processor = NewResourceProcessor(name, config, s.logger)
	default:
		s.logger.Warn("Unknown processor type, using mock", zap.String("name", name))
		processor = &MockProcessor{
			name:   name,
			config: config,
			logger: s.logger,
		}
	}

	s.processors[name] = processor
	return nil
}

// initializeExporter initializes an exporter component
func (s *Service) initializeExporter(name string, config interface{}) error {
	s.logger.Debug("Initializing exporter", zap.String("name", name))

	// For now, create a mock exporter
	// In a real implementation, this would create actual exporter instances
	exporter := &MockExporter{
		name:   name,
		config: config,
		logger: s.logger,
	}

	s.exporters[name] = exporter
	return nil
}
