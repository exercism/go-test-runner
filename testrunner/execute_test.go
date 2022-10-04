package testrunner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

const version = 2

// TestRunTests_broken covers the case the code under test does not compile,
// i.e. "go build ." would fail.
func TestRunTests_broken(t *testing.T) {
	input_dir := "./testdata/practice/broken"
	cmdres, ok := runTests(input_dir)
	if ok {
		t.Errorf("Broken test did not fail %s", cmdres.String())
	}

	res := cmdres.String()
	lines := strings.Split(res, "\n")

	expectedLineSuffixes := []string{
		"# gigasecond [gigasecond.test]",
		"broken.go:11:2: undefined: unknownVar",
		"broken.go:12:2: undefined: UnknownFunction",
		"FAIL	gigasecond [build failed]",
		"returned exit code 2: exit status 2",
	}

	for i, expectedSuffix := range expectedLineSuffixes {
		if !strings.HasSuffix(lines[i], expectedSuffix) {
			t.Errorf("Broken test run - unexpected suffix in line: %s, want: %s", lines[i], expectedSuffix)
		}
	}
}

// TestRunTests_missingFunc covers the case that the test code does not compile,
// i.e. "go build ." would succeed but "go test" returns compilation errors.
func TestRunTests_missingFunc(t *testing.T) {
	input_dir := "./testdata/practice/missing_func"
	cmdres, ok := runTests(input_dir)
	if ok {
		t.Errorf("Missing function test did not fail %s", cmdres.String())
	}

	res := cmdres.String()
	lines := strings.Split(res, "\n")

	expectedLineSuffixes := []string{
		"# gigasecond [gigasecond.test]",
		"missing_func_test.go:39:11: undefined: AddGigasecond",
		"missing_func_test.go:72:11: undefined: AddGigasecond",
		"FAIL	gigasecond [build failed]",
		"returned exit code 2: exit status 2",
	}

	for i, expectedSuffix := range expectedLineSuffixes {
		if !strings.HasSuffix(lines[i], expectedSuffix) {
			t.Errorf("Missing function test run - unexpected suffix in line: %s, want: %s", lines[i], expectedSuffix)
		}
	}
}

func TestRunTests_brokenImport(t *testing.T) {
	input_dir := "./testdata/practice/broken_import"
	cmdres, ok := runTests(input_dir)
	if ok {
		t.Errorf("Broken import test did not fail %s", cmdres.String())
	}

	res := cmdres.String()
	lines := strings.Split(res, "\n")

	expectedLineSuffixes := []string{
		"broken_import.go:5:8: expected ';', found ','",
		"returned exit code 1: exit status 1",
	}

	for i, expectedSuffix := range expectedLineSuffixes {
		if !strings.HasSuffix(lines[i], expectedSuffix) {
			t.Errorf("Broken import test run - unexpected suffix in line: %s, want: %s", lines[i], expectedSuffix)
		}
	}
}

func TestRunTests_RuntimeError(t *testing.T) {
	input_dir := "./testdata/practice/runtime_error"
	cmdres, ok := runTests(input_dir)
	if !ok {
		fmt.Printf("runtime error test expected to return ok: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, version)
	jsonBytes, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		t.Fatalf("runtime error output not valid json: %s", err)
	}

	result := string(jsonBytes)

	pre := `{
	"status": "fail",
	"version": 2,
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
	cmdres, ok := runTests(input_dir)
	if !ok {
		fmt.Printf("race detector test expected to return ok: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, version)
	if output.Status != "fail" {
		t.Errorf("wrong status for race detector test: got %q, want %q", output.Status, "fail")
	}

	if !strings.Contains(output.Tests[0].Message, "WARNING: DATA RACE") {
		t.Errorf("no data race error included in message: %s", output.Tests[0].Message)
	}
}

func TestRunTests_passing(t *testing.T) {
	input_dir := "./testdata/practice/passing"

	cmdres, ok := runTests(input_dir)
	if !ok {
		t.Errorf("Passing test failed: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, version)
	jsonBytes, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		t.Fatalf("Passing test output not valid json: %s", err)
	}

	expectedOutput, err := ioutil.ReadFile("./testdata/practice/passing/output.json")
	if err != nil {
		t.Fatalf("Passing test failed to read test file: %s", err)
	}

	if string(jsonBytes) != string(expectedOutput) {
		t.Errorf("Passing test failed, got:\n%s\n, want:\n%s\n", string(jsonBytes), string(expectedOutput))
	}
}

func TestRunTests_PkgLevelError(t *testing.T) {
	input_dir := "./testdata/practice/pkg_level_error"
	cmdres, ok := runTests(input_dir)
	if !ok {
		fmt.Printf("pkg level error test expected to return ok: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, version)
	jsonBytes, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		t.Fatalf("pkg level error output not valid json: %s", err)
	}

	result := string(jsonBytes)

	pre := `{
	"status": "error",
	"version": 2,
	"message": "panic: Please implement this function`

	if !strings.HasPrefix(result, pre) {
		t.Errorf("pkg level error result has unexpected json prefix: %s", result)
	}
}

func ExampleRunTests_failing() {
	input_dir := "./testdata/practice/failing"

	cmdres, ok := runTests(input_dir)
	if !ok {
		fmt.Printf("Failing test expected to return ok: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, version)
	if bts, err := json.MarshalIndent(output, "", "\t"); err != nil {
		fmt.Printf("Failing test output not valid json: %s", err)
	} else {
		fmt.Println(string(bts))
	}
	// Output: {
	//	"status": "fail",
	//	"version": 2,
	//	"tests": [
	//		{
	//			"name": "TestTrivialFail",
	//			"status": "fail",
	//			"test_code": "// Trivial failing test example\nfunc TestTrivialFail(t *testing.T) {\n\tif false != true {\n\t\tt.Fatal(\"Intentional test failure\")\n\t}\n\tfmt.Println(\"sample failing test output\")\n}",
	//			"message": "\n=== RUN   TestTrivialFail\n\n    failing_test.go:11: Intentional test failure\n\n--- FAIL: TestTrivialFail (0.00s)\n"
	//		}
	//	]
	//}
}
