// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package contract

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader handles loading and validating YAML contracts
type Loader struct {
	contracts []*Contract
	errors    []error
}

// NewLoader creates a new contract loader
func NewLoader() *Loader {
	return &Loader{
		contracts: make([]*Contract, 0),
		errors:    make([]error, 0),
	}
}

// LoadFromPaths loads contracts from the specified file paths or glob patterns
func (l *Loader) LoadFromPaths(paths []string) ([]*Contract, []error) {
	for _, path := range paths {
		if err := l.loadPath(path); err != nil {
			l.errors = append(l.errors, fmt.Errorf("failed to load path %s: %w", path, err))
		}
	}
	return l.contracts, l.errors
}

// loadPath handles a single path (file or glob pattern)
func (l *Loader) loadPath(path string) error {
	// Check if it's a glob pattern
	if strings.Contains(path, "*") {
		return l.loadGlob(path)
	}

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return l.loadDirectory(path)
	}

	// Single file
	return l.loadFile(path)
}

// loadGlob loads contracts from glob patterns
func (l *Loader) loadGlob(pattern string) error {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, match := range matches {
		if err := l.loadFile(match); err != nil {
			l.errors = append(l.errors, fmt.Errorf("failed to load %s: %w", match, err))
		}
	}
	return nil
}

// loadDirectory loads all YAML files from a directory
func (l *Loader) loadDirectory(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".yaml") &&
			!strings.HasSuffix(strings.ToLower(path), ".yml") {
			return nil
		}

		return l.loadFile(path)
	})
}

// loadFile loads a single YAML contract file
func (l *Loader) loadFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contract := &Contract{}
	if err := yaml.Unmarshal(data, contract); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Set the file path for reference
	contract.FilePath = filePath

	// Validate the contract
	if err := l.validateContract(contract); err != nil {
		return fmt.Errorf("contract validation failed: %w", err)
	}

	l.contracts = append(l.contracts, contract)
	return nil
}

// validateContract validates a single contract
func (l *Loader) validateContract(contract *Contract) error {
	var errors []string

	// Required fields
	if contract.Publisher == "" {
		errors = append(errors, "publisher is required")
	}
	if contract.Version == "" {
		errors = append(errors, "version is required")
	}

	// Validate pipeline configuration
	if err := l.validatePipelineConfig(contract); err != nil {
		errors = append(errors, fmt.Sprintf("pipeline configuration validation failed: %v", err))
	}

	// Validate inputs
	if err := l.validateInputs(&contract.Inputs); err != nil {
		errors = append(errors, fmt.Sprintf("inputs validation failed: %v", err))
	}

	// Validate filters
	if err := l.validateFilters(contract.Filters); err != nil {
		errors = append(errors, fmt.Sprintf("filters validation failed: %v", err))
	}

	// Validate matchers
	if err := l.validateMatchers(&contract.Matchers); err != nil {
		errors = append(errors, fmt.Sprintf("matchers validation failed: %v", err))
	}

	// Validate time windows
	if err := l.validateTimeWindows(contract.TimeWindows); err != nil {
		errors = append(errors, fmt.Sprintf("time_windows validation failed: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("contract validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validatePipelineConfig validates the pipeline configuration
func (l *Loader) validatePipelineConfig(contract *Contract) error {
	// Either pipeline or pipeline_selectors must be specified
	if contract.Pipeline == "" && (contract.PipelineSelectors == nil || len(contract.PipelineSelectors.Selectors) == 0) {
		return fmt.Errorf("either pipeline or pipeline_selectors must be specified")
	}

	// If both are specified, pipeline_selectors takes precedence
	if contract.Pipeline != "" && contract.PipelineSelectors != nil && len(contract.PipelineSelectors.Selectors) > 0 {
		// This is allowed but we'll log a warning
		log.Printf("both selector and id were provided - selector being used")
	}

	// Validate pipeline selectors if present
	if contract.PipelineSelectors != nil {
		if err := l.validatePipelineSelectors(contract.PipelineSelectors); err != nil {
			return fmt.Errorf("pipeline selectors validation failed: %w", err)
		}
	}

	return nil
}

// validatePipelineSelectors validates the pipeline selectors
func (l *Loader) validatePipelineSelectors(selectors *PipelineSelectors) error {
	if len(selectors.Selectors) == 0 {
		return fmt.Errorf("at least one selector must be specified")
	}

	for i, selector := range selectors.Selectors {
		if err := l.validatePipelineSelector(selector, i); err != nil {
			return fmt.Errorf("selector %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validatePipelineSelector validates a single pipeline selector
func (l *Loader) validatePipelineSelector(selector PipelineSelector, index int) error {
	if selector.Field == "" {
		return fmt.Errorf("field is required")
	}

	switch selector.Operator {
	case PipelineSelectorOperatorEquals, PipelineSelectorOperatorMatches,
		PipelineSelectorOperatorContains, PipelineSelectorOperatorStartsWith,
		PipelineSelectorOperatorEndsWith:
		if selector.Value == nil {
			return fmt.Errorf("value is required for operator %s", selector.Operator)
		}
	default:
		return fmt.Errorf("invalid operator %s", selector.Operator)
	}

	return nil
}

// validateInputs validates the inputs section
func (l *Loader) validateInputs(inputs *Inputs) error {
	// At least one input type should be specified
	if len(inputs.Traces) == 0 && len(inputs.Metrics) == 0 && len(inputs.Logs) == 0 {
		return fmt.Errorf("at least one input type (traces, metrics, or logs) must be specified")
	}

	// Validate trace inputs
	for i, trace := range inputs.Traces {
		if trace.SpanName == "" {
			return fmt.Errorf("trace input %d: span_name is required", i)
		}
	}

	// Validate metric inputs
	for i, metric := range inputs.Metrics {
		if metric.Name == "" {
			return fmt.Errorf("metric input %d: name is required", i)
		}
		if metric.Value == nil {
			return fmt.Errorf("metric input %d: value is required", i)
		}
	}

	// Validate log inputs
	for i, log := range inputs.Logs {
		if log.Body == "" {
			return fmt.Errorf("log input %d: body is required", i)
		}
	}

	return nil
}

// validateFilters validates the filters section
func (l *Loader) validateFilters(filters []Filter) error {
	for i, filter := range filters {
		if filter.Field == "" {
			return fmt.Errorf("filter %d: field is required", i)
		}

		switch filter.Operator {
		case FilterOperatorEquals, FilterOperatorNotEquals, FilterOperatorMatches,
			FilterOperatorGreaterThan, FilterOperatorLessThan:
			if filter.Value == nil {
				return fmt.Errorf("filter %d: value is required for operator %s", i, filter.Operator)
			}
		case FilterOperatorExists, FilterOperatorNotExists:
			// Value is optional for exists/not_exists
		default:
			return fmt.Errorf("filter %d: invalid operator %s", i, filter.Operator)
		}
	}
	return nil
}

// validateMatchers validates the matchers section
func (l *Loader) validateMatchers(matchers *Matchers) error {
	// At least one matcher type should be specified
	if len(matchers.Traces) == 0 && len(matchers.Metrics) == 0 && len(matchers.Logs) == 0 {
		return fmt.Errorf("at least one matcher type (traces, metrics, or logs) must be specified")
	}

	// Validate trace matchers
	for i, matcher := range matchers.Traces {
		if matcher.SpanName == "" && len(matcher.Attributes) == 0 &&
			matcher.ParentSpan == "" && matcher.ServiceName == "" {
			return fmt.Errorf("trace matcher %d: at least one field must be specified", i)
		}
	}

	// Validate metric matchers
	for i, matcher := range matchers.Metrics {
		if matcher.Name == "" && len(matcher.Labels) == 0 && matcher.Type == "" {
			return fmt.Errorf("metric matcher %d: at least one field must be specified", i)
		}
	}

	// Validate log matchers
	for i, matcher := range matchers.Logs {
		if matcher.Body == "" && len(matcher.Attributes) == 0 && matcher.Severity == "" {
			return fmt.Errorf("log matcher %d: at least one field must be specified", i)
		}
	}

	return nil
}

// validateTimeWindows validates the time windows section
func (l *Loader) validateTimeWindows(windows []TimeWindow) error {
	for i, window := range windows {
		if window.Aggregation == "" {
			return fmt.Errorf("time window %d: aggregation is required", i)
		}
		if window.Duration == "" {
			return fmt.Errorf("time window %d: duration is required", i)
		}
		if window.ExpectedBehavior == "" {
			return fmt.Errorf("time window %d: expected_behavior is required", i)
		}
	}
	return nil
}

// GroupByPublisher groups contracts by publisher
func (l *Loader) GroupByPublisher() map[string][]*Contract {
	groups := make(map[string][]*Contract)
	for _, contract := range l.contracts {
		groups[contract.Publisher] = append(groups[contract.Publisher], contract)
	}
	return groups
}

// GroupByPipeline groups contracts by pipeline
func (l *Loader) GroupByPipeline() map[string][]*Contract {
	groups := make(map[string][]*Contract)
	for _, contract := range l.contracts {
		key := fmt.Sprintf("%s/%s", contract.Publisher, contract.Pipeline)
		groups[key] = append(groups[key], contract)
	}
	return groups
}
