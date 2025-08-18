// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package contract

import (
	"os"
	"testing"
)

func TestLoader_LoadFromPaths(t *testing.T) {
	// Create a temporary test contract file
	testContract := `
publisher: "test-service"
pipeline: "test-pipeline"
version: "1.0"
inputs:
  traces:
    - span_name: "test_operation"
      service_name: "test-service"
      attributes:
        test.key: "test_value"
matchers:
  traces:
    - span_name: "test_operation"
      attributes:
        test.key: "test_value"
`

	tmpFile, err := os.CreateTemp("", "test_contract_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testContract); err != nil {
		t.Fatalf("Failed to write test contract: %v", err)
	}
	tmpFile.Close()

	// Test loading
	loader := NewLoader()
	contracts, errors := loader.LoadFromPaths([]string{tmpFile.Name()})

	if len(errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", errors)
	}

	if len(contracts) != 1 {
		t.Fatalf("Expected 1 contract, got %d", len(contracts))
	}

	contract := contracts[0]
	if contract.Publisher != "test-service" {
		t.Errorf("Expected publisher 'test-service', got '%s'", contract.Publisher)
	}

	if contract.Pipeline != "test-pipeline" {
		t.Errorf("Expected pipeline 'test-pipeline', got '%s'", contract.Pipeline)
	}

	if contract.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", contract.Version)
	}

	if len(contract.Inputs.Traces) != 1 {
		t.Errorf("Expected 1 trace input, got %d", len(contract.Inputs.Traces))
	}

	if len(contract.Matchers.Traces) != 1 {
		t.Errorf("Expected 1 trace matcher, got %d", len(contract.Matchers.Traces))
	}
}

func TestLoader_ValidateContract(t *testing.T) {
	loader := NewLoader()

	// Test valid contract
	validContract := &Contract{
		Publisher: "test-service",
		Pipeline:  "test-pipeline",
		Version:   "1.0",
		Inputs: Inputs{
			Traces: []TraceInput{
				{
					SpanName: "test_operation",
				},
			},
		},
		Matchers: Matchers{
			Traces: []TraceMatcher{
				{
					SpanName: "test_operation",
				},
			},
		},
	}

	if err := loader.validateContract(validContract); err != nil {
		t.Errorf("Expected valid contract, got error: %v", err)
	}

	// Test invalid contract (missing required fields)
	invalidContract := &Contract{
		Publisher: "", // Missing publisher
		Pipeline:  "test-pipeline",
		Version:   "1.0",
	}

	if err := loader.validateContract(invalidContract); err == nil {
		t.Error("Expected error for invalid contract, got none")
	}
}
