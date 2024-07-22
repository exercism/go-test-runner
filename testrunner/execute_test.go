package testrunner

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const version = 3

/*
	Ideally, tests should be performed via running the tool and checking the resulting json
	file matches the expected output. This is done in integration_test.go in the project root.
	Tests should only be added in this file if comparing the full output is not possible for some reason.
*/

func TestRunTests_RuntimeError(t *testing.T) {
	input_dir := filepath.Join("testdata", "practice", "runtime_error")

	cmdres, ok := runTests(input_dir, nil)
	if !ok {
		fmt.Printf("runtime error test expected to return ok: %s", cmdres.String())
	}

	testOutput, err := parseTestOutput(cmdres)
	if err != nil {
		t.Fatalf("parsing test output: %s", err)
	}

	report := getStructureForTestsOk(testOutput, input_dir, version, false)

	jsonBytes, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		t.Fatalf("runtime error output not valid json: %s", err)
	}

	result := string(jsonBytes)

	pre := `{
	"status": "fail",
	"version": 3,
	"tests": [
		{
			"name": "TestAddGigasecond",
			"status": "error",
			"test_code": "func TestAddGigasecond(t *testing.T) {\n\tinput, _ := time.Parse(\"2006-01-02\", \"2011-04-25\")\n\tAddGigasecond(input)\n}",
			"message": "\n=== RUN   TestAddGigasecond\n\nruntime: goroutine stack exceeds`

	if !strings.HasPrefix(result, pre) {
		t.Errorf("runtime error result has unexpected json prefix: %s", result)
	}
}

func TestRunTests_RaceDetector(t *testing.T) {

	input_dir := filepath.Join("testdata", "practice", "race")
	cmdres, ok := runTests(input_dir, []string{"-race"})
	if !ok {
		fmt.Printf("race detector test expected to return ok: %s", cmdres.String())
	}

	testOutput, err := parseTestOutput(cmdres)
	if err != nil {
		t.Errorf("parsing test output: %s", err)
	}

	report := getStructureForTestsOk(testOutput, input_dir, version, false)
	if report.Status != "fail" {
		t.Errorf("wrong status for race detector test: got %q, want %q", report.Status, "fail")
	}

	if !strings.Contains(report.Tests[0].Message, "WARNING: DATA RACE") {
		t.Errorf("no data race error included in message: %s", report.Tests[0].Message)
	}
}

func TestAddNonExecutedTests(t *testing.T) {
	tests := []struct {
		name                string
		inputRootLevelTests []rootLevelTest
		inputResults        []testResult
		expected            []testResult
	}{
		{
			name: "works with no results",
			inputRootLevelTests: []rootLevelTest{
				{name: "TestSomething1"},
				{name: "TestSomething2"},
			},
			inputResults: nil,
			expected: []testResult{
				{
					Name:    "TestSomething1",
					Status:  statErr,
					Message: "This test was not executed.",
				},
				{
					Name:    "TestSomething2",
					Status:  statErr,
					Message: "This test was not executed.",
				},
			},
		},
		{
			name: "adds to end of the results",
			inputRootLevelTests: []rootLevelTest{
				{name: "TestSomething1"},
				{name: "TestSomething2"},
				{name: "TestSomething3"},
				{name: "TestSomething4"},
			},
			inputResults: []testResult{
				{Name: "TestSomething1"},
				{Name: "TestSomething2/subtest1"},
				{Name: "TestSomething2/subtest2"},
			},
			expected: []testResult{
				{Name: "TestSomething1"},
				{Name: "TestSomething2/subtest1"},
				{Name: "TestSomething2/subtest2"},
				{
					Name:    "TestSomething3",
					Status:  statErr,
					Message: "This test was not executed.",
				},
				{
					Name:    "TestSomething4",
					Status:  statErr,
					Message: "This test was not executed.",
				},
			},
		},
		{
			name: "adds to the beginning of the results",
			inputRootLevelTests: []rootLevelTest{
				{name: "TestSomething1"},
				{name: "TestSomething2"},
				{name: "TestSomething3"},
				{name: "TestSomething4"},
			},
			inputResults: []testResult{
				{Name: "TestSomething3"},
				{Name: "TestSomething4/subtest1"},
				{Name: "TestSomething4/subtest2"},
			},
			expected: []testResult{
				{
					Name:    "TestSomething1",
					Status:  statErr,
					Message: "This test was not executed.",
				},
				{
					Name:    "TestSomething2",
					Status:  statErr,
					Message: "This test was not executed.",
				},
				{Name: "TestSomething3"},
				{Name: "TestSomething4/subtest1"},
				{Name: "TestSomething4/subtest2"},
			},
		},
		{
			name: "can add results in the middle",
			inputRootLevelTests: []rootLevelTest{
				{name: "TestSomething1"},
				{name: "TestSomething2"},
				{name: "TestSomething3"},
				{name: "TestSomething4"},
				{name: "TestSomething5"},
				{name: "TestSomething6"},
			},
			inputResults: []testResult{
				{Name: "TestSomething1/subtest1"},
				{Name: "TestSomething1/subtest2"},
				{Name: "TestSomething3"},
				{Name: "TestSomething6"},
			},
			expected: []testResult{
				{Name: "TestSomething1/subtest1"},
				{Name: "TestSomething1/subtest2"},
				{
					Name:    "TestSomething2",
					Status:  statErr,
					Message: "This test was not executed.",
				},
				{Name: "TestSomething3"},
				{
					Name:    "TestSomething4",
					Status:  statErr,
					Message: "This test was not executed.",
				},
				{
					Name:    "TestSomething5",
					Status:  statErr,
					Message: "This test was not executed.",
				},
				{Name: "TestSomething6"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addNonExecutedTests(tt.inputRootLevelTests, tt.inputResults)
			assert.Equal(t, tt.expected, result)
		})
	}
}
