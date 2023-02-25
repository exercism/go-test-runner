package main

import (
	"bytes"
	"go/build"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var regexReplacements = []struct {
	regexp     *regexp.Regexp
	replaceStr string
}{
	{
		// Test duration
		regexp:     regexp.MustCompile(`\\t[0-9]+\.[0-9]+s`),
		replaceStr: "",
	},
	{
		// Test duration in brackets
		regexp:     regexp.MustCompile(`\([0-9]+\.[0-9]+s\)`),
		replaceStr: "",
	},
	{
		// Pointer
		regexp:     regexp.MustCompile(`\+?\(?0x[0-9a-f]+\)?`),
		replaceStr: "",
	},
	{
		// Goroutine
		regexp:     regexp.MustCompile(`goroutine [0-9]+`),
		replaceStr: "goroutine x",
	},
	{
		// Line number
		regexp:     regexp.MustCompile(`\.go:[0-9]+(:[0-9]+)?`),
		replaceStr: ".go",
	},
}

func TestIntegration(t *testing.T) {
	tests := []struct {
		inputDir string
		expected string
	}{
		{
			// This test case covers the case the code under test does not compile,
			// i.e. "go build ." would fail.
			inputDir: "./testrunner/testdata/practice/broken",
			expected: "./testrunner/testdata/expected/broken.json",
		},
		{
			// This test case covers the case that the test code does not compile,
			// i.e. "go build ." would succeed but "go test" returns compilation errors.
			inputDir: "./testrunner/testdata/practice/missing_func",
			expected: "./testrunner/testdata/expected/missing_func.json",
		},
		{
			inputDir: "./testrunner/testdata/practice/broken_import",
			expected: "./testrunner/testdata/expected/broken_import.json",
		},
		{
			inputDir: "./testrunner/testdata/practice/passing",
			expected: "./testrunner/testdata/expected/passing.json",
		},
		{
			inputDir: "./testrunner/testdata/practice/pkg_level_error",
			expected: "./testrunner/testdata/expected/pkg_level_error.json",
		},
		{
			inputDir: "./testrunner/testdata/practice/failing",
			expected: "./testrunner/testdata/expected/failing.json",
		},
		{
			inputDir: "./testrunner/testdata/concept/auto_assigned_task_ids",
			expected: "./testrunner/testdata/expected/auto_assigned_task_ids.json",
		},
		{
			inputDir: "./testrunner/testdata/concept/explicit_task_ids",
			expected: "./testrunner/testdata/expected/explicit_task_ids.json",
		},
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

			resultBytes, err := os.ReadFile("./outdir/results.json")
			require.NoError(t, err, "failed to read results")

			result := string(resultBytes)

			result = strings.ReplaceAll(result, goExe, "PATH_PLACEHOLDER")
			result = strings.ReplaceAll(result, currentDir, "PATH_PLACEHOLDER")
			result = strings.ReplaceAll(result, goRoot, "PATH_PLACEHOLDER")

			for _, replacement := range regexReplacements {
				result = replacement.regexp.ReplaceAllString(result, replacement.replaceStr)
			}

			expected, err := os.ReadFile(tt.expected)
			require.NoError(t, err, "failed to read expected result file")

			assert.JSONEq(t, string(expected), result)
		})
	}
}
