package main

import (
	"fmt"
)

func ExampleBrokenTest() {
	if cmdres, ok := runTests("./testdata/practice/broken"); ok {
		fmt.Printf("Broken test did not fail %s", cmdres.String())
	} else {
		fmt.Println(cmdres.String())
	}
	// Output: FAIL	github.com/exercism/go-test-runner/testdata/practice/broken [build failed]
}
