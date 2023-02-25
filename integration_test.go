package main

import (
	"bytes"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var duration = regexp.MustCompile(`\s[0-9]+\.[0-9]+s`)
var durationInBrackets = regexp.MustCompile(`\([0-9]+\.[0-9]+s\)`)
var pointer = regexp.MustCompile(`\+?\(?0x[0-9a-f]+\)?`)
var goroutine = regexp.MustCompile(`goroutine [0-9]+`)
var lineNumber = regexp.MustCompile(`\.go:[0-9]+`)

func TestIntegration(t *testing.T) {
	tests := []struct {
		inputDir string
		expected string
	}{
		// {
		// 	// This test case covers the case the code under test does not compile,
		// 	// i.e. "go build ." would fail.
		// 	inputDir: "./testrunner/testdata/practice/broken",
		// 	expected: "./testrunner/testdata/expected/broken.json",
		// },
		// {
		// 	// This test case covers the case that the test code does not compile,
		// 	// i.e. "go build ." would succeed but "go test" returns compilation errors.
		// 	inputDir: "./testrunner/testdata/practice/missing_func",
		// 	expected: "./testrunner/testdata/expected/missing_func.json",
		// },
		// {
		// 	inputDir: "./testrunner/testdata/practice/broken_import",
		// 	expected: "./testrunner/testdata/expected/broken_import.json",
		// },
		// {
		// 	inputDir: "./testrunner/testdata/practice/passing",
		// 	expected: "./testrunner/testdata/expected/passing.json",
		// },
		// {
		// 	inputDir: "./testrunner/testdata/practice/pkg_level_error",
		// 	expected: "./testrunner/testdata/expected/pkg_level_error.json",
		// },
		// {
		// 	inputDir: "./testrunner/testdata/practice/failing",
		// 	expected: "./testrunner/testdata/expected/failing.json",
		// },
		// {
		// 	inputDir: "./testrunner/testdata/concept/auto_assigned_task_ids",
		// 	expected: "./testrunner/testdata/expected/auto_assigned_task_ids.json",
		// },
		// {
		// 	inputDir: "./testrunner/testdata/concept/explicit_task_ids",
		// 	expected: "./testrunner/testdata/expected/explicit_task_ids.json",
		// },
		{
			inputDir: "./testrunner/testdata/concept/missing_task_ids",
			expected: "./testrunner/testdata/expected/missing_task_ids.json",
		},
	}

	goExe, err := exec.LookPath("go")
	require.NoError(t, err, "failed to find go executable")

	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		goRoot = build.Default.GOROOT
	}

	currentDir, err := os.Getwd()
	require.NoError(t, err, "failed to determine current directory")

	for _, tt := range tests {
		t.Run(tt.inputDir, func(t *testing.T) {
			err := os.RemoveAll("./outdir")
			require.NoError(t, err, "failed to clean up output directory")

			var stdout, stderr bytes.Buffer
			cmd := &exec.Cmd{
				Path:   goExe,
				Args:   []string{goExe, "run", ".", tt.inputDir, "outdir"},
				Stdout: &stdout,
				Stderr: &stderr,
			}
			err = cmd.Run()
			require.NoErrorf(t, err, "failed to execute test runner: %s %s", stdout.String(), stderr.String())

			result, err := os.ReadFile("./outdir/results.json")
			require.NoError(t, err, "failed to read results")

			sanitizedResult := strings.ReplaceAll(string(result), goExe, "PATH_PLACEHOLDER")
			sanitizedResult = strings.ReplaceAll(sanitizedResult, currentDir, "PATH_PLACEHOLDER")
			sanitizedResult = strings.ReplaceAll(sanitizedResult, goRoot, "PATH_PLACEHOLDER")
			sanitizedResult = duration.ReplaceAllString(sanitizedResult, "")
			sanitizedResult = durationInBrackets.ReplaceAllString(sanitizedResult, "")
			sanitizedResult = pointer.ReplaceAllString(sanitizedResult, "")
			sanitizedResult = goroutine.ReplaceAllString(sanitizedResult, "goroutine x")
			sanitizedResult = lineNumber.ReplaceAllString(sanitizedResult, ".go")

			fmt.Println(sanitizedResult)

			expected, err := os.ReadFile(tt.expected)
			require.NoError(t, err, "failed to read expected result file")

			assert.JSONEq(t, string(expected), sanitizedResult)
		})
	}
}
