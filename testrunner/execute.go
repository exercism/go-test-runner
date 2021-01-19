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
	if cmdres, ok := runTests(input_dir); ok {
		report = getStructure(cmdres, input_dir)
	} else {
		report = &testReport{
			Status:  statErr,
			Message: cmdres.String(),
		}
	}

	bts, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal json from `go test` output: %s", err)
	}
	return bts
}

func getStructure(lines bytes.Buffer, input_dir string) *testReport {
	report := &testReport{
		Status: statPass,
		Tests:  nil,
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

	for _, test := range tests {
		if test == nil {
			// just to be sure we dont get a nil pointer exception
			continue
		}
		if test.Status == statErr {
			report.Status = statErr
		}
		if test.Status == statSkip {
			report.Status = statErr
		}
		if report.Status == statPass && test.Status == statFail {
			report.Status = statFail
		}

		report.Tests = append(report.Tests, *test)
	}

	return report
}

func buildTests(lines bytes.Buffer, input_dir string) (map[string]*testResult, error) {
	var (
		tests       = map[string]*testResult{}
		testFileMap = make(map[string]string)
		failMsg     [][]byte
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
			result := &testResult{
				Name:   line.Test,
				Status: statSkip,
			}
			if len(tc) > 0 {
				result.TestCode = tc
			}
			tests[line.Test] = result
		case "output":
			tests[line.Test].Message += "\n" + line.Output
		case statFail:
			tests[line.Test].Status = statFail
		case statPass:
			tests[line.Test].Status = statPass
		}
	}
	if len(failMsg) != 0 {
		return nil, errors.New(string(bytes.Join(failMsg, []byte{'\n'})))
	}
	return tests, nil
}

// Run the "go test --json ." command, return output
func runTests(input_dir string) (bytes.Buffer, bool) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		log.Fatal("failed to find go executable: ", err)
	}

	var stdout, stderr bytes.Buffer
	testCmd := &exec.Cmd{
		Dir:    input_dir,
		Path:   goExe,
		Args:   []string{goExe, "test", "--json", "."},
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

	switch exc := exitError.ExitCode(); exc {
	case 1:
		// `go test` returns 1 when tests fail, this is fine
		return stdout, true
	case 2:
		//  go test returns 2 on a compilation / build error
		stdout.WriteString(fmt.Sprintf("'%s' returned exit code %d: %s",
			testCmd.String(), exc, err,
		))
		return stdout, false
	default:
		log.Fatalf("error: '%s' failed with exit error %d: %s",
			testCmd.String(), exc, err,
		)
	}
	return stdout, false
}
