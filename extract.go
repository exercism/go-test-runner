package main

import "log"

func extractTestCode(testName string) string {
	//[TODO] determine if testName follows the spec (defined in the README)
	// If so, parse the subtest
	// else, return the entire test function code
	log.Println(testName)
	return testName
}
