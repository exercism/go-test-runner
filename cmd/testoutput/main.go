package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/exercism/go-test-runner/exreport"
)

const (
	statPass = "pass"
	statFail = "fail"
	statSkip = "skip"
	statErr  = "error"
)

func main() {
	lines, err := readStream()
	if err != nil {
		log.Panic(err)
	}

	output := getStructure(lines)
	bts, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(bts))
}

func readStream() ([][]byte, error) {
	_, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	stream, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	return bytes.Split(stream, []byte{'\n'}), nil
}

func getStructure(lines [][]byte) *exreport.Report {
	report := &exreport.Report{
		Status: statPass,
		Tests:  nil,
	}

	tests := buildTests(lines)
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

func buildTests(lines [][]byte) map[string]*exreport.Test {
	var (
		tests   = map[string]*exreport.Test{}
		failMsg [][]byte
	)

forLines:
	for _, lineBytes := range lines {
		var line testLine

		switch {
		case len(lineBytes) == 0:
			continue
		case bytes.HasPrefix(lineBytes, []byte{'#'}):
			// if there is a failure running the tests, supress the line with `#` at the beginning
			continue
		case bytes.HasPrefix(lineBytes, []byte("FAIL")):
			tests["build"] = &exreport.Test{
				Name:    "build",
				Status:  statErr,
				Message: string(bytes.Join(failMsg, []byte{'\n'})),
			}
			break forLines
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
			tests[line.Test] = &exreport.Test{
				Name:   line.Test,
				Status: statSkip,
			}
		case "output":
			tests[line.Test].Message += "\n" + line.Output
		case statFail:
			tests[line.Test].Status = statFail
		case statPass:
			tests[line.Test].Status = statPass
		}
	}
	return tests
}

type testLine struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64
	Output  string
}
