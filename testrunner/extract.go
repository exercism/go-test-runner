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

func FindTestFile(codePath string) string {
	files, err := os.ReadDir(codePath)
	if err != nil {
		log.Printf("warning: input_dir '%s' cannot be read: %s", codePath, err)
		return ""
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), "_test.go") {
			testpath := filepath.Join(codePath, f.Name())
			fh, err := os.ReadFile(testpath)
			if err != nil {
				log.Printf("warning: test file '%s' read failed: %s", testpath, err)
			}

			// We need to check we found the file that actually contains the tests and not only the
			// generated test cases (cases_test.go).
			// Text processing is easier than using AST and should be reliable enough.
			if strings.Contains(string(fh), "func Test") {
				return testpath
			}
		}
	}
	log.Printf("error: test file not found in input_dir '%s'", codePath)
	return ""
}

// return the associated test function code from the given test file
func ExtractTestCodeAndTaskID(rootLevelTests map[string]rootLevelTest, testName string) (string, uint64) {
	test, subtest := splitTestName(testName)
	rootLevelTest := rootLevelTests[test]
	if len(subtest) == 0 {
		return rootLevelTest.code, rootLevelTest.taskID
	}
	defer handleASTPanic()
	subtc := getSubCode(test, subtest, rootLevelTest.code, rootLevelTest.fileName)
	if len(subtc) == 0 {
		return rootLevelTest.code, rootLevelTest.taskID
	}
	return subtc, rootLevelTest.taskID
}

func handleASTPanic() {
	if r := recover(); r != nil {
		fmt.Println("warning: AST parsing failed to extract test code: ", r)
	}
}
