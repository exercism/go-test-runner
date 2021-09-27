package testrunner

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

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
		"# github.com/exercism/go-test-runner/testrunner/testdata/practice/broken [github.com/exercism/go-test-runner/testrunner/testdata/practice/broken.test]",
		"broken.go:11:2: undefined: unknownVar",
		"broken.go:12:2: undefined: UnknownFunction",
		"FAIL	github.com/exercism/go-test-runner/testrunner/testdata/practice/broken [build failed]",
		"returned exit code 2: exit status 2",
	}

	for i, expectedSuffix := range expectedLineSuffixes {
		if !strings.HasSuffix(lines[i], expectedSuffix) {
			t.Errorf("Broken test run - unexpected suffix in line: %s, want: %s", lines[i], expectedSuffix)
		}
	}

	output := &testReport{
		Status:  statErr,
		Version: 2,
		Message: res,
	}
	btr, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		t.Errorf("Broken test output not valid json: %s", err)
	}
	tr := string(btr)

	pre := `{
	"status": "error",
	"version": 2,
	"message": "# github.com/exercism/go-test-runner/testrunner/testdata/practice/broken`

	post := `returned exit code 2: exit status 2",
	"tests": null
}`
	if !strings.HasPrefix(tr, pre) {
		t.Errorf("Broken test run unexpected json prefix: %s", tr)
	}
	if !strings.HasSuffix(tr, post) {
		t.Errorf("Broken test run unexpected json suffix: %s", tr)
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
		"# github.com/exercism/go-test-runner/testrunner/testdata/practice/missing_func [github.com/exercism/go-test-runner/testrunner/testdata/practice/missing_func.test]",
		"missing_func_test.go:39:11: undefined: AddGigasecond",
		"missing_func_test.go:72:11: undefined: AddGigasecond",
		"FAIL	github.com/exercism/go-test-runner/testrunner/testdata/practice/missing_func [build failed]",
		"returned exit code 2: exit status 2",
	}

	for i, expectedSuffix := range expectedLineSuffixes {
		if !strings.HasSuffix(lines[i], expectedSuffix) {
			t.Errorf("Missing function test run - unexpected suffix in line: %s, want: %s", lines[i], expectedSuffix)
		}
	}

	output := &testReport{
		Status:  statErr,
		Version: 2,
		Message: res,
	}
	btr, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		t.Errorf("Missing function test output not valid json: %s", err)
	}
	tr := string(btr)

	pre := `{
	"status": "error",
	"version": 2,
	"message": "# github.com/exercism/go-test-runner/testrunner/testdata/practice/missing_func`

	post := `returned exit code 2: exit status 2",
	"tests": null
}`
	if !strings.HasPrefix(tr, pre) {
		t.Errorf("Missing function test run unexpected json prefix: %s", tr)
	}
	if !strings.HasSuffix(tr, post) {
		t.Errorf("Missing function test run unexpected json suffix: %s", tr)
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

	output := &testReport{
		Status:  statErr,
		Version: 2,
		Message: res,
	}
	btr, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		t.Errorf("Broken import test output not valid json: %s", err)
	}
	tr := string(btr)

	pre := `{
	"status": "error",
	"version": 2,
	"message":`

	post := `returned exit code 1: exit status 1",
	"tests": null
}`
	if !strings.HasPrefix(tr, pre) {
		t.Errorf("Broken import test run unexpected json prefix: %s", tr)
	}
	if !strings.HasSuffix(tr, post) {
		t.Errorf("Broken import test run unexpected json suffix: %s", tr)
	}
}

func ExampleRunTests_passing() {
	input_dir := "./testdata/practice/passing"

	cmdres, ok := runTests(input_dir)
	if !ok {
		fmt.Printf("Passing test failed: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, 2)
	if bts, err := json.MarshalIndent(output, "", "\t"); err != nil {
		fmt.Printf("Passing test output not valid json: %s", err)
	} else {
		fmt.Println(string(bts))
	}
	// Output: {
	//	"status": "pass",
	//	"version": 2,
	//	"tests": [
	//		{
	//			"name": "TestTrivialPass",
	//			"status": "pass",
	//			"test_code": "// Trivial passing test example\nfunc TestTrivialPass(t *testing.T) {\n\tif true != true {\n\t\tt.Fatal(\"This was supposed to be a tautological statement!\")\n\t}\n\tfmt.Println(\"sample passing test output\")\n}",
	//			"message": "\n=== RUN   TestTrivialPass\n\nsample passing test output\n\n--- PASS: TestTrivialPass (0.00s)\n"
	//		}
	//	]
	//}
}

func ExampleRunTests_failing() {
	input_dir := "./testdata/practice/failing"

	cmdres, ok := runTests(input_dir)
	if !ok {
		fmt.Printf("Failing test expected to return ok: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir, 2)
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
