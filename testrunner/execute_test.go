package testrunner

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const version = 3

/*
	Ideally, tests should be performed via running the tool and checking the resulting json
	file matches the expected output. This is done in integration_test.go in the project root.
	Tests should only be added in this file is comparing the full output is not possible for some reason.
*/

func TestRunTests_RuntimeError(t *testing.T) {
	input_dir := "./testdata/practice/runtime_error"
	cmdres, ok := runTests(input_dir, nil)
	if !ok {
		fmt.Printf("runtime error test expected to return ok: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, version, false)
	jsonBytes, err := json.MarshalIndent(output, "", "\t")
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
	input_dir := "./testdata/practice/race"
	cmdres, ok := runTests(input_dir, []string{"-race"})
	if !ok {
		fmt.Printf("race detector test expected to return ok: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, version, false)
	if output.Status != "fail" {
		t.Errorf("wrong status for race detector test: got %q, want %q", output.Status, "fail")
	}

	if !strings.Contains(output.Tests[0].Message, "WARNING: DATA RACE") {
		t.Errorf("no data race error included in message: %s", output.Tests[0].Message)
	}
}
