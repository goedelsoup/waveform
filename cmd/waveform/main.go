// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	if err := rootCmd.MarkFlagRequired("contracts"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTests(cmd *cobra.Command, args []string) error {
	// Load runner configuration
	runnerConfigLoader := config.NewRunnerConfigLoader()
	runnerConfig, err := runnerConfigLoader.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load runner configuration: %w", err)
	}

	// Setup logging based on runner configuration
	var logger *zap.Logger
	if verbose || runnerConfig.Runner.Output.Verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Error syncing logger: %v\n", err)
		}
	}()

	logger.Info("Starting OpenTelemetry Contract Testing Framework",
		zap.String("log_level", runnerConfig.Runner.LogLevel),
		zap.String("log_format", runnerConfig.Runner.LogFormat),
		zap.String("environment", runnerConfig.Global.Environment))

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

	// Generate output files based on runner configuration and command line flags
	outputDir := runnerConfig.Runner.Output.Directory
	if outputDir == "" {
		outputDir = "./waveform-reports"
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("Failed to create output directory", zap.Error(err))
	}

	// Determine which formats to generate
	formatsToGenerate := make(map[string]bool)

	// Add formats from runner configuration
	for _, format := range runnerConfig.Runner.Output.Formats {
		formatsToGenerate[format] = true
	}

	// Override with command line flags
	if junitOutput != "" {
		formatsToGenerate["junit"] = true
	}
	if lcovOutput != "" {
		formatsToGenerate["lcov"] = true
	}
	if summaryOutput != "" {
		formatsToGenerate["summary"] = true
	}

	// Generate reports
	for format := range formatsToGenerate {
		switch format {
		case "junit":
			outputPath := junitOutput
			if outputPath == "" {
				outputPath = filepath.Join(outputDir, "test-results.xml")
			}
			logger.Info("Generating JUnit XML report", zap.String("path", outputPath))
			if err := reportGen.GenerateJUnitXML(outputPath); err != nil {
				logger.Error("Failed to generate JUnit XML report", zap.Error(err))
			}

		case "lcov":
			outputPath := lcovOutput
			if outputPath == "" {
				outputPath = filepath.Join(outputDir, "coverage.info")
			}
			logger.Info("Generating LCOV report", zap.String("path", outputPath))
			if err := reportGen.GenerateLCOV(outputPath); err != nil {
				logger.Error("Failed to generate LCOV report", zap.Error(err))
			}

		case "summary":
			outputPath := summaryOutput
			if outputPath == "" {
				outputPath = filepath.Join(outputDir, "summary.txt")
			}
			logger.Info("Generating summary report", zap.String("path", outputPath))
			if err := reportGen.GenerateSummary(outputPath); err != nil {
				logger.Error("Failed to generate summary report", zap.Error(err))
			}
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
