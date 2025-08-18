// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package main

import (
	"fmt"
	"log"

	"github.com/goedelsoup/waveform/internal/contract"
)

func main() {
	// Create a pipeline selector service
	selectorService := contract.NewPipelineSelectorService()

	// Register some example pipelines
	pipelines := []*contract.PipelineInfo{
		{
			ID:          "trace-auth-prod",
			Name:        "Auth Service Trace Pipeline",
			Description: "Processes trace data from auth service in production",
			Type:        "trace",
			Tags: map[string]string{
				"environment": "production",
				"service":     "auth",
				"datacenter":  "us-east-1",
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
				"datacenter":  "us-west-2",
			},
		},
		{
			ID:          "trace-payment-prod",
			Name:        "Payment Service Trace Pipeline",
			Description: "Processes trace data from payment service in production",
			Type:        "trace",
			Tags: map[string]string{
				"environment": "production",
				"service":     "payment",
				"datacenter":  "us-east-1",
			},
			Metadata: map[string]string{
				"processing_type": "security",
			},
		},
	}

	selectorService.RegisterPipelines(pipelines)

	fmt.Println("Registered Pipelines:")
	fmt.Println("=====================")
	for _, pipeline := range selectorService.ListPipelines() {
		fmt.Printf("- %s (%s): %s\n", pipeline.Name, pipeline.ID, pipeline.Description)
	}
	fmt.Println()

	// Example 1: Find production trace pipelines
	fmt.Println("Example 1: Production Trace Pipelines")
	fmt.Println("=====================================")
	selectors1 := &contract.PipelineSelectors{
		Selectors: []contract.PipelineSelector{
			{
				Field:    "type",
				Operator: contract.PipelineSelectorOperatorEquals,
				Value:    "trace",
			},
			{
				Field:    "tags.environment",
				Operator: contract.PipelineSelectorOperatorEquals,
				Value:    "production",
			},
		},
	}

	matches1, err := selectorService.FindMatchingPipelines(selectors1)
	if err != nil {
		log.Fatalf("Error finding matches: %v", err)
	}

	fmt.Printf("Found %d matching pipelines:\n", len(matches1))
	for _, match := range matches1 {
		fmt.Printf("- %s (%s)\n", match.Name, match.ID)
	}
	fmt.Println()

	// Example 2: Find auth service pipelines
	fmt.Println("Example 2: Auth Service Pipelines")
	fmt.Println("=================================")
	selectors2 := &contract.PipelineSelectors{
		Selectors: []contract.PipelineSelector{
			{
				Field:    "name",
				Operator: contract.PipelineSelectorOperatorContains,
				Value:    "Auth",
			},
		},
	}

	matches2, err := selectorService.FindMatchingPipelines(selectors2)
	if err != nil {
		log.Fatalf("Error finding matches: %v", err)
	}

	fmt.Printf("Found %d matching pipelines:\n", len(matches2))
	for _, match := range matches2 {
		fmt.Printf("- %s (%s)\n", match.Name, match.ID)
	}
	fmt.Println()

	// Example 3: Find aggregation pipelines
	fmt.Println("Example 3: Aggregation Pipelines")
	fmt.Println("=================================")
	selectors3 := &contract.PipelineSelectors{
		Selectors: []contract.PipelineSelector{
			{
				Field:    "metadata.processing_type",
				Operator: contract.PipelineSelectorOperatorEquals,
				Value:    "aggregation",
			},
		},
	}

	matches3, err := selectorService.FindMatchingPipelines(selectors3)
	if err != nil {
		log.Fatalf("Error finding matches: %v", err)
	}

	fmt.Printf("Found %d matching pipelines:\n", len(matches3))
	for _, match := range matches3 {
		fmt.Printf("- %s (%s)\n", match.Name, match.ID)
	}
	fmt.Println()

	// Example 4: Find US East pipelines
	fmt.Println("Example 4: US East Datacenter Pipelines")
	fmt.Println("=======================================")
	selectors4 := &contract.PipelineSelectors{
		Selectors: []contract.PipelineSelector{
			{
				Field:    "tags.datacenter",
				Operator: contract.PipelineSelectorOperatorStartsWith,
				Value:    "us-east",
			},
		},
	}

	matches4, err := selectorService.FindMatchingPipelines(selectors4)
	if err != nil {
		log.Fatalf("Error finding matches: %v", err)
	}

	fmt.Printf("Found %d matching pipelines:\n", len(matches4))
	for _, match := range matches4 {
		fmt.Printf("- %s (%s)\n", match.Name, match.ID)
	}
	fmt.Println()

	// Example 5: Complex pattern matching
	fmt.Println("Example 5: Complex Pattern Matching")
	fmt.Println("===================================")
	selectors5 := &contract.PipelineSelectors{
		Selectors: []contract.PipelineSelector{
			{
				Field:    "name",
				Operator: contract.PipelineSelectorOperatorMatches,
				Value:    "(?i).*service.*pipeline",
			},
			{
				Field:    "tags.environment",
				Operator: contract.PipelineSelectorOperatorEquals,
				Value:    "production",
			},
		},
	}

	matches5, err := selectorService.FindMatchingPipelines(selectors5)
	if err != nil {
		log.Fatalf("Error finding matches: %v", err)
	}

	fmt.Printf("Found %d matching pipelines:\n", len(matches5))
	for _, match := range matches5 {
		fmt.Printf("- %s (%s)\n", match.Name, match.ID)
	}
	fmt.Println()

	// Example 6: Find best match
	fmt.Println("Example 6: Best Match for Trace Processing")
	fmt.Println("==========================================")
	selectors6 := &contract.PipelineSelectors{
		Selectors: []contract.PipelineSelector{
			{
				Field:    "type",
				Operator: contract.PipelineSelectorOperatorEquals,
				Value:    "trace",
			},
		},
		Priority: 1,
	}

	bestMatch, err := selectorService.FindBestMatchingPipeline(selectors6)
	if err != nil {
		log.Fatalf("Error finding best match: %v", err)
	}

	fmt.Printf("Best match: %s (%s)\n", bestMatch.Name, bestMatch.ID)
	fmt.Printf("Description: %s\n", bestMatch.Description)
}
