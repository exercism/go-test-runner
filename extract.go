package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
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
	//[TODO] this should use parser.ParseExprFrom instead of string parsing
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
	// [TODO] properly escape the return string
	test, subtest := splitTestName(testName)
	if 0 == len(subtest) {
		return extractFunc(test, testFile)
	}
	subtcode := extractSub(test, subtest, testFile)
	if 0 == len(subtcode) {
		return extractFunc(test, testFile)
	}
	return subtcode
}

// return the code defining the argument function
func extractFunc(testName string, testFile string) string {
	fset := token.NewFileSet()
	ppc := parser.ParseComments
	if file, err := parser.ParseFile(fset, testFile, nil, ppc); err == nil {
		for _, d := range file.Decls {
			if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == testName {
				fun := &printer.CommentedNode{Node: f, Comments: file.Comments}
				var buf bytes.Buffer
				printer.Fprint(&buf, fset, fun)
				return buf.String()
			}
		}
	} else {
		log.Printf(
			"warning: '%s' not parsed from '%s': %s", testName, testFile, err,
		)
	}
	return ""
}

func extractSub(test string, sub string, file string) string {
	return sub
}
