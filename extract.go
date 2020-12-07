package main

import (
	"strings"
)

func splitTestName(testName string) (string, string) {
	t := strings.Split(testName, "/")
	if 1 == len(t) {
		return t[0], ""
	}
	return t[0], t[1]
}

func extractTestCode(testName string, input_dir string) string {
	//[TODO] determine if testName follows the spec (defined in the README)
	// If so, parse the subtest
	// else, return the entire test function code
	//log.Println(testName)
	return testName
}
