// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package contract

import (
	"fmt"
	"regexp"
	"strings"
)

// PipelineInfo represents information about a pipeline that can be matched against selectors
type PipelineInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Type        string            `json:"type,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PipelineSelectorService handles matching contracts to pipelines based on selectors
type PipelineSelectorService struct {
	pipelines map[string]*PipelineInfo
}

// NewPipelineSelectorService creates a new pipeline selector service
func NewPipelineSelectorService() *PipelineSelectorService {
	return &PipelineSelectorService{
		pipelines: make(map[string]*PipelineInfo),
	}
}

// RegisterPipeline registers a pipeline with the selector service
func (s *PipelineSelectorService) RegisterPipeline(pipeline *PipelineInfo) {
	s.pipelines[pipeline.ID] = pipeline
}

// RegisterPipelines registers multiple pipelines with the selector service
func (s *PipelineSelectorService) RegisterPipelines(pipelines []*PipelineInfo) {
	for _, pipeline := range pipelines {
		s.RegisterPipeline(pipeline)
	}
}

// FindMatchingPipelines finds all pipelines that match the given selectors
func (s *PipelineSelectorService) FindMatchingPipelines(selectors *PipelineSelectors) ([]*PipelineInfo, error) {
	if selectors == nil || len(selectors.Selectors) == 0 {
		return nil, fmt.Errorf("no selectors provided")
	}

	var matchingPipelines []*PipelineInfo

	for _, pipeline := range s.pipelines {
		if s.matchesSelectors(pipeline, selectors.Selectors) {
			matchingPipelines = append(matchingPipelines, pipeline)
		}
	}

	return matchingPipelines, nil
}

// FindBestMatchingPipeline finds the best matching pipeline for the given selectors
func (s *PipelineSelectorService) FindBestMatchingPipeline(selectors *PipelineSelectors) (*PipelineInfo, error) {
	matches, err := s.FindMatchingPipelines(selectors)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matching pipelines found")
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	// If multiple matches, return the first one
	// In a real implementation, you might want to implement more sophisticated ranking
	return matches[0], nil
}

// matchesSelectors checks if a pipeline matches all the given selectors
func (s *PipelineSelectorService) matchesSelectors(pipeline *PipelineInfo, selectors []PipelineSelector) bool {
	for _, selector := range selectors {
		if !s.matchesSelector(pipeline, selector) {
			return false
		}
	}
	return true
}

// matchesSelector checks if a pipeline matches a single selector
func (s *PipelineSelectorService) matchesSelector(pipeline *PipelineInfo, selector PipelineSelector) bool {
	fieldValue := s.extractFieldValue(pipeline, selector.Field)
	if fieldValue == nil {
		return false
	}

	switch selector.Operator {
	case PipelineSelectorOperatorEquals:
		return s.compareValues(fieldValue, selector.Value, "equals")
	case PipelineSelectorOperatorMatches:
		return s.matchesPattern(fieldValue, selector.Value)
	case PipelineSelectorOperatorContains:
		return s.containsValue(fieldValue, selector.Value)
	case PipelineSelectorOperatorStartsWith:
		return s.startsWithValue(fieldValue, selector.Value)
	case PipelineSelectorOperatorEndsWith:
		return s.endsWithValue(fieldValue, selector.Value)
	default:
		return false
	}
}

// extractFieldValue extracts a field value from pipeline info based on a dot-separated path
func (s *PipelineSelectorService) extractFieldValue(pipeline *PipelineInfo, fieldPath string) interface{} {
	parts := strings.Split(fieldPath, ".")
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "id":
		return pipeline.ID
	case "name":
		return pipeline.Name
	case "description":
		return pipeline.Description
	case "type":
		return pipeline.Type
	case "tags":
		if len(parts) > 1 {
			return pipeline.Tags[parts[1]]
		}
		return nil
	case "metadata":
		if len(parts) > 1 {
			return pipeline.Metadata[parts[1]]
		}
		return nil
	default:
		return nil
	}
}

// compareValues compares two values for equality
func (s *PipelineSelectorService) compareValues(a, b interface{}, operation string) bool {
	if a == nil || b == nil {
		return a == b
	}

	// Convert to strings for comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	switch operation {
	case "equals":
		return aStr == bStr
	default:
		return false
	}
}

// matchesPattern checks if a value matches a regex pattern
func (s *PipelineSelectorService) matchesPattern(value interface{}, pattern interface{}) bool {
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

// containsValue checks if a value contains another value
func (s *PipelineSelectorService) containsValue(value interface{}, contains interface{}) bool {
	if value == nil || contains == nil {
		return false
	}

	valueStr, ok := value.(string)
	if !ok {
		valueStr = fmt.Sprintf("%v", value)
	}

	containsStr, ok := contains.(string)
	if !ok {
		containsStr = fmt.Sprintf("%v", contains)
	}

	return strings.Contains(valueStr, containsStr)
}

// startsWithValue checks if a value starts with another value
func (s *PipelineSelectorService) startsWithValue(value interface{}, startsWith interface{}) bool {
	if value == nil || startsWith == nil {
		return false
	}

	valueStr, ok := value.(string)
	if !ok {
		valueStr = fmt.Sprintf("%v", value)
	}

	startsWithStr, ok := startsWith.(string)
	if !ok {
		startsWithStr = fmt.Sprintf("%v", startsWith)
	}

	return strings.HasPrefix(valueStr, startsWithStr)
}

// endsWithValue checks if a value ends with another value
func (s *PipelineSelectorService) endsWithValue(value interface{}, endsWith interface{}) bool {
	if value == nil || endsWith == nil {
		return false
	}

	valueStr, ok := value.(string)
	if !ok {
		valueStr = fmt.Sprintf("%v", value)
	}

	endsWithStr, ok := endsWith.(string)
	if !ok {
		endsWithStr = fmt.Sprintf("%v", endsWith)
	}

	return strings.HasSuffix(valueStr, endsWithStr)
}

// GetPipelineByID returns a pipeline by its ID
func (s *PipelineSelectorService) GetPipelineByID(id string) (*PipelineInfo, error) {
	pipeline, exists := s.pipelines[id]
	if !exists {
		return nil, fmt.Errorf("pipeline with ID %s not found", id)
	}
	return pipeline, nil
}

// ListPipelines returns all registered pipelines
func (s *PipelineSelectorService) ListPipelines() []*PipelineInfo {
	pipelines := make([]*PipelineInfo, 0, len(s.pipelines))
	for _, pipeline := range s.pipelines {
		pipelines = append(pipelines, pipeline)
	}
	return pipelines
}
