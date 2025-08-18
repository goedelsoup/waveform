// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package report

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goedelsoup/waveform/internal/harness"
)

// JUnitTestSuite represents a JUnit test suite
type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      float64         `xml:"time,attr"`
	Timestamp string          `xml:"timestamp,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase represents a JUnit test case
type JUnitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
	Error     *JUnitError   `xml:"error,omitempty"`
}

// JUnitFailure represents a JUnit test failure
type JUnitFailure struct {
	XMLName xml.Name `xml:"failure"`
	Message string   `xml:"message,attr"`
	Type    string   `xml:"type,attr"`
	Content string   `xml:",chardata"`
}

// JUnitError represents a JUnit test error
type JUnitError struct {
	XMLName xml.Name `xml:"error"`
	Message string   `xml:"message,attr"`
	Type    string   `xml:"type,attr"`
	Content string   `xml:",chardata"`
}

// LCOVRecord represents an LCOV coverage record
type LCOVRecord struct {
	TestName     string
	Publisher    string
	Pipeline     string
	Covered      bool
	Duration     time.Duration
	ErrorCount   int
	WarningCount int
}

// ReportGenerator generates test reports in various formats
type ReportGenerator struct {
	results harness.TestResults
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(results harness.TestResults) *ReportGenerator {
	return &ReportGenerator{
		results: results,
	}
}

// GenerateJUnitXML generates a JUnit XML report
func (r *ReportGenerator) GenerateJUnitXML(outputPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create test suite
	testSuite := JUnitTestSuite{
		Name:      "OpenTelemetry Contract Tests",
		Tests:     r.results.TotalTests,
		Failures:  r.results.FailedTests,
		Errors:    0, // We don't distinguish between failures and errors for now
		Skipped:   0,
		Time:      r.results.Duration.Seconds(),
		Timestamp: time.Now().Format(time.RFC3339),
		TestCases: make([]JUnitTestCase, 0, len(r.results.Results)),
	}

	// Add test cases
	for _, result := range r.results.Results {
		testCase := JUnitTestCase{
			Name:      fmt.Sprintf("%s/%s", result.Contract.Publisher, result.Contract.Pipeline),
			Classname: result.Contract.Publisher,
			Time:      result.Duration.Seconds(),
		}

		if !result.Valid {
			// Create failure message
			failureMessage := "Contract validation failed"
			if len(result.Errors) > 0 {
				failureMessage = result.Errors[0]
			}

			testCase.Failure = &JUnitFailure{
				Message: failureMessage,
				Type:    "ValidationError",
				Content: r.formatErrors(result.Errors),
			}
		}

		testSuite.TestCases = append(testSuite.TestCases, testCase)
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(testSuite, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML header
	xmlData = append([]byte(xml.Header), xmlData...)

	// Write to file
	if err := os.WriteFile(outputPath, xmlData, 0644); err != nil {
		return fmt.Errorf("failed to write XML file: %w", err)
	}

	return nil
}

// GenerateLCOV generates an LCOV format report
func (r *ReportGenerator) GenerateLCOV(outputPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create LCOV records
	records := make([]LCOVRecord, 0, len(r.results.Results))
	for _, result := range r.results.Results {
		record := LCOVRecord{
			TestName:     fmt.Sprintf("%s/%s", result.Contract.Publisher, result.Contract.Pipeline),
			Publisher:    result.Contract.Publisher,
			Pipeline:     result.Contract.Pipeline,
			Covered:      result.Valid,
			Duration:     result.Duration,
			ErrorCount:   len(result.Errors),
			WarningCount: len(result.Warnings),
		}
		records = append(records, record)
	}

	// Generate LCOV content
	lcovContent := r.generateLCOVContent(records)

	// Write to file
	if err := os.WriteFile(outputPath, []byte(lcovContent), 0644); err != nil {
		return fmt.Errorf("failed to write LCOV file: %w", err)
	}

	return nil
}

// GenerateSummary generates a human-readable summary report
func (r *ReportGenerator) GenerateSummary(outputPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Generate summary content
	summary := r.generateSummaryContent()

	// Write to file
	if err := os.WriteFile(outputPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	return nil
}

// formatErrors formats error messages for XML output
func (r *ReportGenerator) formatErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}

	result := "Errors:\n"
	for i, err := range errors {
		result += fmt.Sprintf("  %d. %s\n", i+1, err)
	}
	return result
}

// generateLCOVContent generates LCOV format content
func (r *ReportGenerator) generateLCOVContent(records []LCOVRecord) string {
	content := "# LCOV coverage report for OpenTelemetry Contract Tests\n"
	content += fmt.Sprintf("# Generated at: %s\n", time.Now().Format(time.RFC3339))
	content += fmt.Sprintf("# Total tests: %d\n", len(records))

	passed := 0
	for _, record := range records {
		if record.Covered {
			passed++
		}
	}
	content += fmt.Sprintf("# Passed tests: %d\n", passed)
	content += fmt.Sprintf("# Failed tests: %d\n", len(records)-passed)
	content += fmt.Sprintf("# Coverage: %.2f%%\n", float64(passed)/float64(len(records))*100)
	content += "\n"

	// Add test details
	for _, record := range records {
		status := "PASS"
		if !record.Covered {
			status = "FAIL"
		}
		content += fmt.Sprintf("TN:%s\n", record.TestName)
		content += fmt.Sprintf("TF:%s\n", record.Publisher)
		content += fmt.Sprintf("FNF:%s\n", record.Pipeline)
		content += fmt.Sprintf("FNH:%s\n", status)
		content += fmt.Sprintf("DA:%d,%d\n", 1, record.ErrorCount)
		content += fmt.Sprintf("LF:%d\n", 1)
		coverage := 0
		if record.Covered {
			coverage = 1
		}
		content += fmt.Sprintf("LH:%d\n", coverage)
		content += "end_of_record\n"
	}

	return content
}

// generateSummaryContent generates a human-readable summary
func (r *ReportGenerator) generateSummaryContent() string {
	content := "OpenTelemetry Contract Testing Summary\n"
	content += "=====================================\n\n"
	content += fmt.Sprintf("Generated at: %s\n", time.Now().Format(time.RFC3339))
	content += fmt.Sprintf("Total duration: %s\n", r.results.Duration)
	content += fmt.Sprintf("Total tests: %d\n", r.results.TotalTests)
	content += fmt.Sprintf("Passed tests: %d\n", r.results.PassedTests)
	content += fmt.Sprintf("Failed tests: %d\n", r.results.FailedTests)

	if r.results.TotalTests > 0 {
		passRate := float64(r.results.PassedTests) / float64(r.results.TotalTests) * 100
		content += fmt.Sprintf("Pass rate: %.2f%%\n", passRate)
	}

	content += "\nTest Results:\n"
	content += "=============\n\n"

	// Group by publisher
	publisherGroups := make(map[string][]harness.TestResult)
	for _, result := range r.results.Results {
		publisherGroups[result.Contract.Publisher] = append(publisherGroups[result.Contract.Publisher], result)
	}

	for publisher, results := range publisherGroups {
		content += fmt.Sprintf("Publisher: %s\n", publisher)
		content += fmt.Sprintf("  Tests: %d\n", len(results))

		passed := 0
		for _, result := range results {
			if result.Valid {
				passed++
			}
		}
		content += fmt.Sprintf("  Passed: %d\n", passed)
		content += fmt.Sprintf("  Failed: %d\n", len(results)-passed)

		if len(results) > 0 {
			passRate := float64(passed) / float64(len(results)) * 100
			content += fmt.Sprintf("  Pass rate: %.2f%%\n", passRate)
		}

		content += "\n  Details:\n"
		for _, result := range results {
			status := "PASS"
			if !result.Valid {
				status = "FAIL"
			}
			content += fmt.Sprintf("    %s/%s: %s (%s)\n",
				result.Contract.Publisher,
				result.Contract.Pipeline,
				status,
				result.Duration)

			if !result.Valid && len(result.Errors) > 0 {
				content += fmt.Sprintf("      Error: %s\n", result.Errors[0])
			}
		}
		content += "\n"
	}

	return content
}

// PrintSummary prints a summary to stdout
func (r *ReportGenerator) PrintSummary() {
	summary := r.generateSummaryContent()
	fmt.Print(summary)
}
