package testrunner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"
)

const (
	statPass = "pass"
	statFail = "fail"
	statSkip = "skip"
	statErr  = "error"
)

type testResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	TestCode string `json:"test_code"`
	Message  string `json:"message"`
}

type testReport struct {
	Status  string       `json:"status"`
	Version int          `json:"version"`
	Message string       `json:"message,omitempty"`
	Tests   []testResult `json:"tests"`
}

type testLine struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64
	Output  string
}

func Execute(input_dir string) []byte {
	var report *testReport
	ver := 2
	if cmdres, ok := runTests(input_dir); ok {
		report = getStructure(cmdres, input_dir, ver)
	} else {
		report = &testReport{
			Status:  statErr,
			Version: ver,
			Message: cmdres.String(),
		}
	}

	bts, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal json from `go test` output: %s", err)
	}
	return bts
}

func getStructure(lines bytes.Buffer, input_dir string, ver int) *testReport {
	report := &testReport{
		Status:  statPass,
		Version: ver,
		Tests:   nil,
	}
	defer func() {
		if report.Tests == nil {
			report.Tests = []testResult{}
		}
	}()

	tests, err := buildTests(lines, input_dir)
	if err != nil {
		report.Status = statErr
		report.Message = err.Error()
		return report
	}

	tests = removeObsoleteParentTests(tests)

	for _, test := range tests {
		if test.Status == statSkip {
			// There is no status for skipped tests on the website
			// so we remove them from the output.
			continue
		}
		if test.Status == statErr {
			report.Status = statErr
		}
		if report.Status == statPass && test.Status == statFail {
			report.Status = statFail
		}

		report.Tests = append(report.Tests, test)
	}

	return report
}

func buildTests(lines bytes.Buffer, input_dir string) ([]testResult, error) {
	var (
		results         = []testResult{}
		resultIdxByName = make(map[string]int)
		testFileMap     = make(map[string]string)
		failMsg         [][]byte
	)

	scanner := bufio.NewScanner(&lines)
	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		var line testLine

		switch {
		case len(lineBytes) == 0:
			continue
		case !bytes.HasPrefix(lineBytes, []byte{'{'}):
			// if the line is not a json, we need to collect the lines to gather why `go test --json` failed
			failMsg = append(failMsg, lineBytes)
			continue
		}

		if err := json.Unmarshal(lineBytes, &line); err != nil {
			log.Println(err)
			continue
		}

		if line.Test == "" {
			continue
		}

		switch line.Action {
		case "run":
			tf, cached := testFileMap[line.Test]
			if !cached {
				tf = findTestFile(line.Test, input_dir)
				testFileMap[line.Test] = tf
			}
			tc := ExtractTestCode(line.Test, tf)
			result := testResult{
				Name:   line.Test,
				Status: statErr,
			}
			if len(tc) > 0 {
				result.TestCode = tc
			}

			results = append(results, result)
			resultIdxByName[result.Name] = len(results) - 1
		case "output":
			if idx, found := resultIdxByName[line.Test]; found {
				results[idx].Message += "\n" + line.Output
			} else {
				log.Printf("cannot extend message for unknown test: %s\n", line.Test)
				continue
			}
		case statFail:
			if idx, found := resultIdxByName[line.Test]; found {
				results[idx].Status = statFail
			} else {
				log.Printf("cannot set failed status for unknown test: %s\n", line.Test)
				continue
			}
		case statPass:
			if idx, found := resultIdxByName[line.Test]; found {
				results[idx].Status = statPass
			} else {
				log.Printf("cannot set passing status for unknown test: %s\n", line.Test)
				continue
			}
		case statSkip:
			if idx, found := resultIdxByName[line.Test]; found {
				results[idx].Status = statSkip
			} else {
				log.Printf("cannot set skipped status for unknown test: %s\n", line.Test)
				continue
			}
		}

	}
	if len(failMsg) != 0 {
		return nil, errors.New(string(bytes.Join(failMsg, []byte{'\n'})))
	}
	return results, nil
}

// removeObsoleteParentTests cleans up the list of test results. The parent test
// would just repeat the same code that is shown for the sub tests but would not
// contain the result of the assertions. This is confusing for students. So if a
// sub-test is found, the corresponding parent test is removed from the results.
func removeObsoleteParentTests(tests []testResult) []testResult {
	namesOfObsoleteTests := map[string]bool{}
	for _, test := range tests {
		parentName, subTestName := splitTestName(test.Name)
		if subTestName != "" {
			namesOfObsoleteTests[parentName] = true
		}
	}

	results := []testResult{}
	for _, test := range tests {
		if !namesOfObsoleteTests[test.Name] {
			results = append(results, test)
		}
	}

	return results
}

// codeCompiles runs "go build ." and return whether it worked or not
func codeCompiles(input_dir string) bool {
	goExe, err := exec.LookPath("go")
	if err != nil {
		log.Fatal("failed to find go executable: ", err)
	}

	var stdout, stderr bytes.Buffer
	testCmd := &exec.Cmd{
		Dir:    input_dir,
		Path:   goExe,
		Args:   []string{goExe, "build", "."},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err = testCmd.Run()
	return err == nil
}

// testCompiles compiles the tests and return whether it worked or not
func testCompiles(input_dir string) bool {
	goExe, err := exec.LookPath("go")
	if err != nil {
		log.Fatal("failed to find go executable: ", err)
	}

	var stdout, stderr bytes.Buffer
	testCmd := &exec.Cmd{
		Dir:  input_dir,
		Path: goExe,
		// "Official" recommendation for compiling but not running the tests
		// https://github.com/golang/go/issues/46712#issuecomment-859949958
		Args:   []string{goExe, "test", "-c", "-o", "/dev/null"},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err = testCmd.Run()
	return err == nil
}

// Run the "go test --short --json ." command, return output
// --short is used to exclude benchmark tests, given the spec / web UI currently cannot handle them
func runTests(input_dir string) (bytes.Buffer, bool) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		log.Fatal("failed to find go executable: ", err)
	}

	var stdout, stderr bytes.Buffer
	testCmd := &exec.Cmd{
		Dir:    input_dir,
		Path:   goExe,
		Args:   []string{goExe, "test", "--short", "--json", "."},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err = testCmd.Run()
	if err == nil {
		// Test ran without any problems, return json
		return stdout, true
	}

	exitError, ok := err.(*exec.ExitError)
	if !ok {
		log.Fatalf("error: '%s' failed with non exit error %s",
			testCmd.String(), err,
		)
	}
	exc := exitError.ExitCode()

	// Do the code and the test even compile?
	if !codeCompiles(input_dir) || !testCompiles(input_dir) {
		// Combine stderr and stdout in the same order in which they
		// show up in the console.
		stderr.WriteString(stdout.String())
		stderr.WriteString(fmt.Sprintf("'%s' returned exit code %d: %s",
			testCmd.String(), exc, err,
		))
		return stderr, false
	}

	switch exc {
	case 1:
		// `go test` returns 1 when tests fail, this is fine
		return stdout, true
	default:
		log.Fatalf("error: '%s' failed with exit error %d: %s",
			testCmd.String(), exc, err,
		)
	}
	return stdout, false
}
