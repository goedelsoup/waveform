// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package main

import (
	"fmt"
	"os"

	"github.com/goedelsoup/waveform/internal/config"
	"github.com/goedelsoup/waveform/internal/contract"
	"github.com/goedelsoup/waveform/internal/harness"
	"github.com/goedelsoup/waveform/internal/report"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	contractPaths []string
	testMode      string
	configPath    string
	junitOutput   string
	lcovOutput    string
	summaryOutput string
	verbose       bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "waveform",
		Short: "OpenTelemetry Contract Testing Framework",
		Long: `A standalone Go testing framework that applies contract testing principles 
to OpenTelemetry pipelines. The framework allows telemetry publishers to define 
YAML contracts specifying their expectations, then validates that collector 
pipelines transform data correctly.`,
		RunE: runTests,
	}

	// Add flags
	rootCmd.Flags().StringSliceVarP(&contractPaths, "contracts", "c", []string{}, "Contract file paths or glob patterns")
	rootCmd.Flags().StringVarP(&testMode, "mode", "m", "pipeline", "Test mode: pipeline or processor")
	rootCmd.Flags().StringVarP(&configPath, "config", "f", "", "Collector configuration file path")
	rootCmd.Flags().StringVarP(&junitOutput, "junit-output", "j", "", "JUnit XML output file path")
	rootCmd.Flags().StringVarP(&lcovOutput, "lcov-output", "l", "", "LCOV output file path")
	rootCmd.Flags().StringVarP(&summaryOutput, "summary-output", "s", "", "Summary output file path")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	// Mark required flags
	rootCmd.MarkFlagRequired("contracts")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTests(cmd *cobra.Command, args []string) error {
	// Setup logging
	var logger *zap.Logger
	var err error
	if verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting OpenTelemetry Contract Testing Framework")

	// Load contracts
	logger.Info("Loading contracts", zap.Strings("paths", contractPaths))
	loader := contract.NewLoader()
	contracts, errors := loader.LoadFromPaths(contractPaths)

	if len(errors) > 0 {
		logger.Warn("Some contracts failed to load", zap.Int("error_count", len(errors)))
		for _, err := range errors {
			logger.Error("Contract loading error", zap.Error(err))
		}
	}

	if len(contracts) == 0 {
		return fmt.Errorf("no valid contracts found")
	}

	logger.Info("Contracts loaded successfully", zap.Int("count", len(contracts)))

	// Load collector configuration if provided
	var collectorConfig harness.CollectorConfig
	if configPath != "" {
		logger.Info("Loading collector configuration", zap.String("path", configPath))

		// Create configuration loader
		configLoader := config.NewLoader()

		// Load configuration from file
		loadedConfig, err := configLoader.LoadFromFile(configPath)
		if err != nil {
			logger.Error("Failed to load collector configuration", zap.Error(err))
			return fmt.Errorf("failed to load collector configuration: %w", err)
		}

		// Validate configuration
		if err := configLoader.ValidateConfig(loadedConfig); err != nil {
			logger.Error("Invalid collector configuration", zap.Error(err))
			return fmt.Errorf("invalid collector configuration: %w", err)
		}

		collectorConfig = *loadedConfig
		logger.Info("Collector configuration loaded successfully",
			zap.Int("receivers", len(collectorConfig.Receivers)),
			zap.Int("processors", len(collectorConfig.Processors)),
			zap.Int("exporters", len(collectorConfig.Exporters)))
	} else {
		// Use empty configuration if no config file provided
		collectorConfig = harness.CollectorConfig{
			Receivers:  make(map[string]interface{}),
			Processors: make(map[string]interface{}),
			Exporters:  make(map[string]interface{}),
			Service:    make(map[string]interface{}),
		}
	}

	// Create test harness
	mode := harness.TestMode(testMode)
	harness := harness.NewTestHarness(mode, collectorConfig)
	harness.SetLogger(logger)

	// Run tests
	logger.Info("Running tests", zap.String("mode", string(mode)))
	results := harness.RunTests(contracts)

	// Generate reports
	reportGen := report.NewReportGenerator(results)

	// Print summary to stdout
	reportGen.PrintSummary()

	// Generate output files
	if junitOutput != "" {
		logger.Info("Generating JUnit XML report", zap.String("path", junitOutput))
		if err := reportGen.GenerateJUnitXML(junitOutput); err != nil {
			logger.Error("Failed to generate JUnit XML report", zap.Error(err))
		}
	}

	if lcovOutput != "" {
		logger.Info("Generating LCOV report", zap.String("path", lcovOutput))
		if err := reportGen.GenerateLCOV(lcovOutput); err != nil {
			logger.Error("Failed to generate LCOV report", zap.Error(err))
		}
	}

	if summaryOutput != "" {
		logger.Info("Generating summary report", zap.String("path", summaryOutput))
		if err := reportGen.GenerateSummary(summaryOutput); err != nil {
			logger.Error("Failed to generate summary report", zap.Error(err))
		}
	}

	// Exit with appropriate code
	if results.FailedTests > 0 {
		logger.Info("Tests completed with failures",
			zap.Int("total", results.TotalTests),
			zap.Int("passed", results.PassedTests),
			zap.Int("failed", results.FailedTests))
		os.Exit(1)
	}

	logger.Info("All tests passed",
		zap.Int("total", results.TotalTests),
		zap.Duration("duration", results.Duration))
	return nil
}
