// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package contract

import (
	"testing"
)

func TestPipelineSelectorService(t *testing.T) {
	service := NewPipelineSelectorService()

	// Register some test pipelines
	pipelines := []*PipelineInfo{
		{
			ID:          "trace-auth-prod",
			Name:        "Auth Service Trace Pipeline",
			Description: "Processes trace data from auth service in production",
			Type:        "trace",
			Tags: map[string]string{
				"environment": "production",
				"service":     "auth",
			},
			Metadata: map[string]string{
				"processing_type": "validation",
			},
		},
		{
			ID:          "metric-user-prod",
			Name:        "User Service Metric Pipeline",
			Description: "Processes metric data from user service in production",
			Type:        "metric",
			Tags: map[string]string{
				"environment": "production",
				"service":     "user",
				"datacenter":  "us-east-1",
			},
			Metadata: map[string]string{
				"processing_type": "aggregation",
			},
		},
		{
			ID:          "log-auth-staging",
			Name:        "Auth Service Log Pipeline",
			Description: "Processes log data from auth service in staging",
			Type:        "log",
			Tags: map[string]string{
				"environment": "staging",
				"service":     "auth",
			},
		},
	}

	service.RegisterPipelines(pipelines)

	t.Run("Test Exact Match", func(t *testing.T) {
		selectors := &PipelineSelectors{
			Selectors: []PipelineSelector{
				{
					Field:    "type",
					Operator: PipelineSelectorOperatorEquals,
					Value:    "trace",
				},
				{
					Field:    "tags.environment",
					Operator: PipelineSelectorOperatorEquals,
					Value:    "production",
				},
			},
		}

		matches, err := service.FindMatchingPipelines(selectors)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(matches) != 1 {
			t.Fatalf("Expected 1 match, got %d", len(matches))
		}

		if matches[0].ID != "trace-auth-prod" {
			t.Errorf("Expected trace-auth-prod, got %s", matches[0].ID)
		}
	})

	t.Run("Test Pattern Match", func(t *testing.T) {
		selectors := &PipelineSelectors{
			Selectors: []PipelineSelector{
				{
					Field:    "name",
					Operator: PipelineSelectorOperatorMatches,
					Value:    "(?i).*auth.*",
				},
			},
		}

		matches, err := service.FindMatchingPipelines(selectors)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(matches) != 2 {
			t.Fatalf("Expected 2 matches, got %d", len(matches))
		}

		// Check that both auth pipelines are matched
		found := make(map[string]bool)
		for _, match := range matches {
			found[match.ID] = true
		}

		if !found["trace-auth-prod"] || !found["log-auth-staging"] {
			t.Error("Expected both auth pipelines to be matched")
		}
	})

	t.Run("Test Contains Match", func(t *testing.T) {
		selectors := &PipelineSelectors{
			Selectors: []PipelineSelector{
				{
					Field:    "name",
					Operator: PipelineSelectorOperatorContains,
					Value:    "User",
				},
			},
		}

		matches, err := service.FindMatchingPipelines(selectors)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(matches) != 1 {
			t.Fatalf("Expected 1 match, got %d", len(matches))
		}

		if matches[0].ID != "metric-user-prod" {
			t.Errorf("Expected metric-user-prod, got %s", matches[0].ID)
		}
	})

	t.Run("Test Starts With Match", func(t *testing.T) {
		selectors := &PipelineSelectors{
			Selectors: []PipelineSelector{
				{
					Field:    "tags.datacenter",
					Operator: PipelineSelectorOperatorStartsWith,
					Value:    "us-",
				},
			},
		}

		matches, err := service.FindMatchingPipelines(selectors)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(matches) != 1 {
			t.Fatalf("Expected 1 match, got %d", len(matches))
		}

		if matches[0].ID != "metric-user-prod" {
			t.Errorf("Expected metric-user-prod, got %s", matches[0].ID)
		}
	})

	t.Run("Test Multiple Selectors", func(t *testing.T) {
		selectors := &PipelineSelectors{
			Selectors: []PipelineSelector{
				{
					Field:    "type",
					Operator: PipelineSelectorOperatorEquals,
					Value:    "metric",
				},
				{
					Field:    "metadata.processing_type",
					Operator: PipelineSelectorOperatorEquals,
					Value:    "aggregation",
				},
			},
		}

		matches, err := service.FindMatchingPipelines(selectors)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(matches) != 1 {
			t.Fatalf("Expected 1 match, got %d", len(matches))
		}

		if matches[0].ID != "metric-user-prod" {
			t.Errorf("Expected metric-user-prod, got %s", matches[0].ID)
		}
	})

	t.Run("Test No Match", func(t *testing.T) {
		selectors := &PipelineSelectors{
			Selectors: []PipelineSelector{
				{
					Field:    "type",
					Operator: PipelineSelectorOperatorEquals,
					Value:    "nonexistent",
				},
			},
		}

		matches, err := service.FindMatchingPipelines(selectors)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(matches) != 0 {
			t.Fatalf("Expected 0 matches, got %d", len(matches))
		}
	})

	t.Run("Test Best Match", func(t *testing.T) {
		selectors := &PipelineSelectors{
			Selectors: []PipelineSelector{
				{
					Field:    "type",
					Operator: PipelineSelectorOperatorEquals,
					Value:    "trace",
				},
			},
		}

		match, err := service.FindBestMatchingPipeline(selectors)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if match.ID != "trace-auth-prod" {
			t.Errorf("Expected trace-auth-prod, got %s", match.ID)
		}
	})
}
