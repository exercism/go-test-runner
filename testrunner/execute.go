package testrunner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	statPass = "pass"
	statFail = "fail"
	statSkip = "skip"
	statErr  = "error"
)

// For security reasons, only testing flags that are included in the list below are processed.
var allowedTestingFlags = []string{"-race"}

type testResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	TestCode string `json:"test_code"`
	Message  string `json:"message"`
	TaskID   uint64 `json:"task_id,omitempty"`
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
	ver := 3

	exerciseConfig := parseExerciseConfig(input_dir)
	cmdres, testsOk := runTests(input_dir, exerciseConfig.TestingFlags)
	testOutput, err := parseTestOutput(cmdres)
	if err != nil {
		log.Fatalf("parsing test output: %s", err)
	}

	if testsOk {
		report = getStructureForTestsOk(testOutput, input_dir, ver, exerciseConfig.TaskIDsEnabled)
	} else {
		report = getStructureForTestsNotOk(testOutput, ver)
	}

	bts, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal json from `go test` output: %s", err)
	}
	return bts
}

func getStructureForTestsNotOk(parsedOutput *parsedTestOutput, ver int) *testReport {
	report := &testReport{
		Status:  statErr,
		Version: ver,
		Message: parsedOutput.joinPackageMessages("\n"),
	}

	report.Message += parsedOutput.joinFailMessages("\n")

	jsonOutputMessages := make([]string, 0)

	for _, line := range parsedOutput.testLines {
		if line.Action == "output" {
			jsonOutputMessages = append(jsonOutputMessages, line.Output)
		}
	}

	if len(jsonOutputMessages) > 0 {
		report.Message += "\n" + strings.Join(jsonOutputMessages, "\n")
	}

	return report
}

func getStructureForTestsOk(parsedOutput *parsedTestOutput, input_dir string, ver int, taskIDsEnabled bool) *testReport {
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

	tests := processTestResults(parsedOutput, input_dir, taskIDsEnabled)

	if parsedOutput.hasFailMessages() {
		report.Status = statErr
		report.Message = parsedOutput.joinFailMessages("\n")
		return report
	}

	if len(tests) == 0 && parsedOutput.hasPackageMessages() {
		report.Status = statErr
		report.Message = parsedOutput.joinPackageMessages("")
		return report
	}

	tests = removeObsoleteParentTests(tests)
	tests = formatTestNames(tests)
	tests = cleanUpTaskIDs(tests, taskIDsEnabled)

	for _, test := range tests {
		if test.Status == statSkip {
			// There is no status for skipped tests on the website
			// so we remove them from the output.
			continue
		}
		if test.Status == statErr {
			// If only one test has an error, the overall report should
			// only say "fail". Report level "error" is only for cases
			// when we don't have any test level results.
			report.Status = statFail
		}
		if report.Status == statPass && test.Status == statFail {
			report.Status = statFail
		}

		report.Tests = append(report.Tests, test)
	}

	return report
}

type parsedTestOutput struct {
	testLines        []testLine
	pkgLevelMessages []string
	failMessages     []string
}

func (out *parsedTestOutput) hasFailMessages() bool {
	return len(out.failMessages) > 0
}

func (out *parsedTestOutput) joinFailMessages(sep string) string {
	return strings.Join(out.failMessages, sep)
}

func (out *parsedTestOutput) hasPackageMessages() bool {
	return len(out.pkgLevelMessages) > 0
}

func (out *parsedTestOutput) joinPackageMessages(sep string) string {
	return strings.Join(out.pkgLevelMessages, sep)
}

func parseTestOutput(lines bytes.Buffer) (*parsedTestOutput, error) {
	parsedOutput := &parsedTestOutput{
		testLines:        make([]testLine, 0),
		pkgLevelMessages: make([]string, 0),
		failMessages:     make([]string, 0),
	}

	scanner := bufio.NewScanner(&lines)
	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		var line testLine

		if len(lineBytes) == 0 {
			continue
		}

		if !bytes.HasPrefix(lineBytes, []byte{'{'}) {
			// if the line is not a json, we need to collect the lines to gather why `go test --json` failed
			parsedOutput.failMessages = append(parsedOutput.failMessages, string(lineBytes))
			continue
		}

		if err := json.Unmarshal(lineBytes, &line); err != nil {
			return nil, fmt.Errorf("parsing line starting with '{' as json: %w", err)
		}

		if line.Test == "" {
			// We collect messages that do not belong to an individual test and use them later
			// as error message in case there was no test level message found at all.
			if line.Output != "" {
				parsedOutput.pkgLevelMessages = append(parsedOutput.pkgLevelMessages, line.Output)
			}
			continue
		}

		parsedOutput.testLines = append(parsedOutput.testLines, line)
	}

	return parsedOutput, nil
}

func processTestResults(
	parsedOutput *parsedTestOutput,
	input_dir string,
	taskIDsEnabled bool,
) []testResult {

	results := make([]testResult, 0)
	resultIdxByName := make(map[string]int)

	testFile := FindTestFile(input_dir)
	rootLevelTests := FindAllRootLevelTests(testFile)
	rootLevelTestsMap := ConvertToMapByTestName(rootLevelTests)

	for _, parsedLine := range parsedOutput.testLines {
		switch parsedLine.Action {
		case "run":
			tc, taskID := ExtractTestCodeAndTaskID(rootLevelTestsMap, parsedLine.Test)
			result := testResult{
				Name: parsedLine.Test,
				// Use error as default state in case no other state is found later.
				// No state is provided e.g. when there is a stack overflow.
				Status: statErr,
				TaskID: taskID,
			}
			if len(tc) > 0 {
				result.TestCode = tc
			}

			results = append(results, result)
			resultIdxByName[result.Name] = len(results) - 1
		case "output":
			if idx, found := resultIdxByName[parsedLine.Test]; found {
				results[idx].Message += "\n" + parsedLine.Output
			} else {
				log.Printf("cannot extend message for unknown test: %s\n", parsedLine.Test)
				continue
			}
		case statFail:
			if idx, found := resultIdxByName[parsedLine.Test]; found {
				results[idx].Status = statFail
			} else {
				log.Printf("cannot set failed status for unknown test: %s\n", parsedLine.Test)
				continue
			}
		case statPass:
			if idx, found := resultIdxByName[parsedLine.Test]; found {
				results[idx].Status = statPass
			} else {
				log.Printf("cannot set passing status for unknown test: %s\n", parsedLine.Test)
				continue
			}
		case statSkip:
			if idx, found := resultIdxByName[parsedLine.Test]; found {
				results[idx].Status = statSkip
			} else {
				log.Printf("cannot set skipped status for unknown test: %s\n", parsedLine.Test)
				continue
			}
		}
	}

	if taskIDsEnabled {
		// We only need this for the V3 UI with task ids.
		// It causes issues for some practice exercises.
		results = addNonExecutedTests(rootLevelTests, results)
	}

	return results
}

// addNonExecutedTests adds tests to the result set that were not executed.
// They are added with status "error" and special message (this is common in other tracks as well).
// The function makes sure that the result for non-executed test is inserted in the correct position.
func addNonExecutedTests(rootLevelTests []rootLevelTest, results []testResult) []testResult {
	insertResultAfterIdx := -1
	for parentIdx, parentTest := range rootLevelTests {
		parentFound := false

		// For the given parent test, check whether the result set already contains
		// test results for it.
		for resultIdx := range results {
			parentName, _ := splitTestName(results[resultIdx].Name)
			if rootLevelTests[parentIdx].name == parentName {
				insertResultAfterIdx = resultIdx
				parentFound = true
				// No "continue" here, we need to find the index of the last (sub)test result
				// that belongs to a given parent test name.
			}
		}

		// If we found test results for the parent test name,
		// there is nothing to do.
		if parentFound {
			continue
		}

		// If not, we insert the new test result for the test that was not executed.
		newResult := testResult{
			Name:     parentTest.name,
			Status:   statErr,
			TestCode: parentTest.code,
			Message:  "This test was not executed.",
		}

		if insertResultAfterIdx < 0 {
			results = append([]testResult{newResult}, results...)
		} else if insertResultAfterIdx >= len(results)-1 {
			results = append(results, newResult)
		} else {
			secondPart := append([]testResult{newResult}, results[insertResultAfterIdx+1:]...)
			results = append(results[:insertResultAfterIdx+1], secondPart...)
		}
		insertResultAfterIdx++
	}

	return results
}

var parentTestMsg = regexp.MustCompile(`(?s)=== RUN\s*Test.*--- (?:FAIL|PASS): Test.*? \(.*?\)\s(.*)`)

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

	// If the parent test includes a message (besides the standard "RUN ... PASS/FAIL ...")
	// we keep that message in a map.
	testNameToMsg := map[string]string{}
	results := []testResult{}
	for _, test := range tests {
		if !namesOfObsoleteTests[test.Name] {
			results = append(results, test)
		} else {
			match := parentTestMsg.FindStringSubmatch(test.Message)
			if len(match) == 2 && len(strings.TrimSpace(match[1])) > 0 {
				testNameToMsg[test.Name] = match[1]
			}
		}
	}

	// We add the message we found on the parent to the first subtest
	// for that parent.
	for i, test := range results {
		parentName, _ := splitTestName(test.Name)
		if testNameToMsg[parentName] != "" {
			results[i].Message += testNameToMsg[parentName]
			delete(testNameToMsg, parentName)
		}
	}

	return results
}

// formatTestNames makes sure the test names contain spaces so that
// line breaks are possible on the website. With that, the test names
// are readable even if the sidebar with the test results is narrow.
func formatTestNames(tests []testResult) []testResult {
	out := make([]testResult, 0, len(tests))
	replacer := strings.NewReplacer("/", "/ ", "_", " ")
	for _, test := range tests {
		test.Name = replacer.Replace(test.Name)
		out = append(out, test)
	}
	return out
}

// cleanUpTaskIDs makes sure no task IDs are set when they were not enabled
// in the config. If task ids are enabled and explicit values were found for all
// parent tests, those are kept.
// If no explicit task IDs where found, it will assign incrementing
// task IDs to all tests in the list.
func cleanUpTaskIDs(tests []testResult, taskIDsEnabled bool) []testResult {
	if len(tests) == 0 {
		return tests
	}

	if !taskIDsEnabled {
		for i := range tests {
			tests[i].TaskID = 0
		}
		return tests
	}

	allTestsHaveTaskIDs := true
	taskIDSeen := false
	for _, test := range tests {
		if test.TaskID == 0 {
			allTestsHaveTaskIDs = false
		} else {
			taskIDSeen = true
		}
	}

	if allTestsHaveTaskIDs {
		return tests
	}

	// If we found some task ids but not all of them, we remove all task ids to be safe.
	if taskIDSeen {
		for i := range tests {
			tests[i].TaskID = 0
		}
		return tests
	}

	// No explicit task IDs found, performing auto-assignment.
	currentParent := ""
	currentTaskID := uint64(0)
	for i := range tests {
		parentName, _ := splitTestName(tests[i].Name)
		if parentName != currentParent {
			currentParent = parentName
			// Only increment the number, if a new parent test starts.
			currentTaskID++
		}
		tests[i].TaskID = currentTaskID
	}

	return tests
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
		Args:   []string{goExe, "test", "-c", "-o", os.DevNull},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err = testCmd.Run()
	return err == nil
}

// Run the "go test --short --json ." command, return output
// --short is used to exclude benchmark tests, given the spec / web UI currently cannot handle them
func runTests(input_dir string, additionalTestFlags []string) (bytes.Buffer, bool) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		log.Fatal("failed to find go executable: ", err)
	}

	testCommand := []string{goExe, "test", "--short", "--json"}
	testCommand = append(testCommand, additionalTestFlags...)
	testCommand = append(testCommand, ".")

	var stdout, stderr bytes.Buffer
	testCmd := &exec.Cmd{
		Dir:    input_dir,
		Path:   goExe,
		Args:   testCommand,
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

type config struct {
	Custom ExerciseConfig `json:"custom"`
}

type ExerciseConfig struct {
	TestingFlags   []string `json:"testingFlags"`
	TaskIDsEnabled bool     `json:"taskIdsEnabled"`
}

func parseExerciseConfig(input_dir string) ExerciseConfig {
	configContent, err := os.ReadFile(filepath.Join(input_dir, ".meta", "config.json"))
	if err != nil {
		log.Printf("warning: config.json could not be read: %v", err)
		return ExerciseConfig{}
	}

	cfg := &config{}
	err = json.Unmarshal(configContent, cfg)
	if err != nil {
		log.Printf("failed to parse config.json: %v", err)
		return ExerciseConfig{}
	}

	if len(cfg.Custom.TestingFlags) != 0 {
		cfg.Custom.TestingFlags = validateTestingFlags(cfg.Custom.TestingFlags)
	}

	return cfg.Custom
}

func validateTestingFlags(flags []string) []string {
	var validFlags []string
	for _, flag := range flags {
		if contains(allowedTestingFlags, flag) {
			validFlags = append(validFlags, flag)
		} else {
			log.Printf("invalid testing flag found in config.json: %s", flag)
		}
	}
	return validFlags
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}
