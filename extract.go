package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// Split a test name into its constituent parts
// https://blog.golang.org/subtests#:~:text=The%20full%20name%20of%20a,first%20argument%20to%20Run%20otherwise.
func splitTestName(testName string) (string, string) {
	t := strings.Split(testName, "/")
	if 1 == len(t) {
		return t[0], ""
	}
	return t[0], t[1]
}

// Search a code path and return the file containing the test argument
func findTestFile(testName string, codePath string) string {
	test, _ := splitTestName(testName)
	files, err := ioutil.ReadDir(codePath)
	if err != nil {
		log.Printf("warning: input_dir '%s' cannot be read: %s", codePath, err)
		return ""
	}
	testdef := fmt.Sprintf("func %s", test)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "_test.go") {
			var code string
			testpath := filepath.Join(codePath, f.Name())
			fmt.Scanln(&code)
			fh, err := ioutil.ReadFile(testpath)
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
func extractTestCode(testName string, testFile string) string {
	test, subtest := splitTestName(testName)
	tc := getFuncCode(test, testFile)
	if 0 == len(subtest) {
		return tc
	}
	subtc := getSubCode(test, subtest, tc, testFile)
	if 0 == len(subtc) {
		return tc
	}
	return subtc
}
