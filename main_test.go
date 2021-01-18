package gotestrun

import (
	"encoding/json"
	"fmt"
)

func ExampleBrokenTestRun() {
	input_dir := "./testdata/practice/broken"
	if cmdres, ok := runTests(input_dir); ok {
		fmt.Printf("Broken test did not fail %s", cmdres.String())
	} else {
		fmt.Println(cmdres.String())
	}
	// Output: FAIL	github.com/exercism/go-test-runner/testdata/practice/broken [build failed]
	//'/usr/local/bin/go test --json .' returned exit code 2: exit status 2
}

func ExampleBrokenTestJson() {
	input_dir := "./testdata/practice/broken"
	cmdres, ok := runTests(input_dir)
	if ok {
		fmt.Printf("Broken test did not fail: %s", cmdres.String())
	}
	output := &testReport{
		Status:  statErr,
		Message: cmdres.String(),
	}
	if bts, err := json.MarshalIndent(output, "", "\t"); err != nil {
		fmt.Printf("Broken test output not valid json: %s", err)
	} else {
		fmt.Println(string(bts))
	}
	// Output: {
	//	"status": "error",
	//	"message": "FAIL\tgithub.com/exercism/go-test-runner/testdata/practice/broken [build failed]\n'/usr/local/bin/go test --json .' returned exit code 2: exit status 2",
	//	"tests": null
	//}
}

func ExamplePassingTestJson() {
	input_dir := "./testdata/practice/passing"

	cmdres, ok := runTests(input_dir)
	if !ok {
		fmt.Printf("Passing test failed: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir)
	if bts, err := json.MarshalIndent(output, "", "\t"); err != nil {
		fmt.Printf("Passing test output not valid json: %s", err)
	} else {
		fmt.Println(string(bts))
	}
	// Output: {
	//	"status": "pass",
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

func ExampleFailingTestJson() {
	input_dir := "./testdata/practice/failing"

	cmdres, ok := runTests(input_dir)
	if !ok {
		fmt.Printf("Failing test expected to return ok: %s", cmdres.String())
	}

	output := getStructure(cmdres, input_dir)
	if bts, err := json.MarshalIndent(output, "", "\t"); err != nil {
		fmt.Printf("Failing test output not valid json: %s", err)
	} else {
		fmt.Println(string(bts))
	}
	// Output: {
	//	"status": "fail",
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
