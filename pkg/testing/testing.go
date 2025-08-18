// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package testing

import (
	"fmt"
	"time"

	"github.com/goedelsoup/waveform/internal/contract"
	"github.com/goedelsoup/waveform/internal/harness"
	"github.com/goedelsoup/waveform/internal/report"
	"go.uber.org/zap"
)

// Framework provides the main API for the OpenTelemetry Contract Testing Framework
type Framework struct {
	harness *harness.TestHarness
	logger  *zap.Logger
}

// NewFramework creates a new testing framework instance
func NewFramework(mode harness.TestMode, config harness.CollectorConfig) *Framework {
	return &Framework{
		harness: harness.NewTestHarness(mode, config),
		logger:  zap.NewNop(),
	}
}

// SetLogger sets the logger for the framework
func (f *Framework) SetLogger(logger *zap.Logger) {
	f.logger = logger
	f.harness.SetLogger(logger)
}

// RunTests runs tests for the given contracts
func (f *Framework) RunTests(contracts []*contract.Contract) harness.TestResults {
	return f.harness.RunTests(contracts)
}

// GenerateReport creates a report generator for the given test results
func (f *Framework) GenerateReport(results harness.TestResults) *report.ReportGenerator {
	return report.NewReportGenerator(results)
}

// TestOptions provides options for running tests
type TestOptions struct {
	IgnoreTimestamps bool
	TimeTolerance    time.Duration
	Verbose          bool
}

// DefaultTestOptions returns default test options
func DefaultTestOptions() *TestOptions {
	return &TestOptions{
		IgnoreTimestamps: true,
		TimeTolerance:    1 * time.Second,
		Verbose:          false,
	}
}

// RunContractTests runs tests for contracts with the given options
func RunContractTests(contracts []*contract.Contract, mode harness.TestMode, config harness.CollectorConfig, options *TestOptions) (harness.TestResults, error) {
	if options == nil {
		options = DefaultTestOptions()
	}

	// Setup logging
	var logger *zap.Logger
	var err error
	if options.Verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return harness.TestResults{}, err
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			// Log the error but don't fail the test
			fmt.Printf("Error syncing logger: %v\n", err)
		}
	}()

	// Create framework
	framework := NewFramework(mode, config)
	framework.SetLogger(logger)

	// Run tests
	results := framework.RunTests(contracts)

	return results, nil
}

// LoadContracts loads contracts from file paths
func LoadContracts(paths []string) ([]*contract.Contract, []error) {
	loader := contract.NewLoader()
	return loader.LoadFromPaths(paths)
}

// GroupContractsByPublisher groups contracts by publisher
func GroupContractsByPublisher(contracts []*contract.Contract) map[string][]*contract.Contract {
	groups := make(map[string][]*contract.Contract)
	for _, contract := range contracts {
		groups[contract.Publisher] = append(groups[contract.Publisher], contract)
	}
	return groups
}

// GroupContractsByPipeline groups contracts by pipeline
func GroupContractsByPipeline(contracts []*contract.Contract) map[string][]*contract.Contract {
	groups := make(map[string][]*contract.Contract)
	for _, contract := range contracts {
		key := contract.Publisher + "/" + contract.Pipeline
		groups[key] = append(groups[key], contract)
	}
	return groups
}
