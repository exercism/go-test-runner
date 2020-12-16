package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: go-test-runner input_dir output_dir")
	}

	input_dir := os.Args[1]
	output_dir := os.Args[2]

	if _, err := os.Stat(input_dir); os.IsNotExist(err) {
		log.Fatal("input_dir does not exist: ", input_dir)
	}

	if _, err := os.Stat(output_dir); os.IsNotExist(err) {
		log.Printf(
			"output_dir does not exist, attempting to create: %s", output_dir,
		)
		if err := os.Mkdir(output_dir, os.ModeDir|0755); err != nil {
			log.Fatal("output_dir does not exist: ", output_dir)
		}
	}

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

	if err := testCmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if 1 == exitError.ExitCode() {
				// `go test` returns 1 when tests fail
				// The test runner should continue and return 0 in this case
				log.Printf(
					"warning: ignoring exit code 1 from '%s'", testCmd.String(),
				)
			} else {
				log.Fatalf("'%s' failed with exit code %d: %s",
					testCmd.String(), exitError.ExitCode(), err)
			}
		}
	}

	output := getStructure(stdout, input_dir)
	bts, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal json from `go test` output: %s", err)
	}

	results := filepath.Join(output_dir, "results.json")
	err = ioutil.WriteFile(results, bts, 0644)
	if err != nil {
		log.Fatalf("Failed to write results.json: %s", err)
	}
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
			tc := extractTestCode(line.Test, tf)
			result := &testResult{
				Name:     line.Test,
				Status:   statSkip,
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
