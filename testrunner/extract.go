package testrunner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Split a test name into its constituent parts
// https://blog.golang.org/subtests#:~:text=The%20full%20name%20of%20a,first%20argument%20to%20Run%20otherwise.
func splitTestName(testName string) (string, string) {
	t := strings.Split(testName, "/")
	if len(t) == 1 {
		return t[0], ""
	}
	return t[0], t[1]
}

// Search a code path and return the file containing the test argument
func findTestFile(testName string, codePath string) string {
	test, _ := splitTestName(testName)
	files, err := os.ReadDir(codePath)
	if err != nil {
		log.Printf("warning: input_dir '%s' cannot be read: %s", codePath, err)
		return ""
	}
	testdef := fmt.Sprintf("func %s", test)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "_test.go") {
			testpath := filepath.Join(codePath, f.Name())
			fh, err := os.ReadFile(testpath)
			if err != nil {
				log.Printf("warning: test file '%s' read failed: %s", testpath, err)
			}
			// Text processing is easier than using AST and should be reliable enough
			if strings.Contains(string(fh), testdef) {
				return testpath
			}
		}
	}
	log.Printf("warning: test %s not found in input_dir '%s'", codePath, test)
	return ""
}

// return the associated test function code from the given test file
func ExtractTestCodeAndTaskID(testName string, testFile string) (string, uint64) {
	test, subtest := splitTestName(testName)
	tc, taskID := getFuncCodeAndTaskID(test, testFile)
	if len(subtest) == 0 {
		return tc, taskID
	}
	defer handleASTPanic()
	subtc := getSubCode(test, subtest, tc, testFile)
	if len(subtc) == 0 {
		return tc, taskID
	}
	return subtc, taskID
}

func handleASTPanic() {
	if r := recover(); r != nil {
		fmt.Println("warning: AST parsing failed to extract test code: ", r)
	}
}
