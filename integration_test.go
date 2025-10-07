package main

import (
	"bytes"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
			inputDir: filepath.Join("testrunner", "testdata", "practice", "broken"),
			expected: filepath.Join("testrunner", "testdata", "expected", "broken.json"),
		},
		{
			// This test case covers the case that the test code does not compile,
			// i.e. "go build ." would succeed but "go test" returns compilation errors.
			inputDir: filepath.Join("testrunner", "testdata", "practice", "missing_func"),
			expected: filepath.Join("testrunner", "testdata", "expected", "missing_func.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "practice", "broken_import"),
			expected: filepath.Join("testrunner", "testdata", "expected", "broken_import.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "practice", "passing"),
			expected: filepath.Join("testrunner", "testdata", "expected", "passing.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "practice", "pkg_level_error"),
			expected: filepath.Join("testrunner", "testdata", "expected", "pkg_level_error.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "practice", "separate_cases_file"),
			expected: filepath.Join("testrunner", "testdata", "expected", "separate_cases_file.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "practice", "failing"),
			expected: filepath.Join("testrunner", "testdata", "expected", "failing.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "concept", "auto_assigned_task_ids"),
			expected: filepath.Join("testrunner", "testdata", "expected", "auto_assigned_task_ids.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "concept", "explicit_task_ids"),
			expected: filepath.Join("testrunner", "testdata", "expected", "explicit_task_ids.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "concept", "missing_task_ids"),
			expected: filepath.Join("testrunner", "testdata", "expected", "missing_task_ids.json"),
		},
		{
			inputDir: filepath.Join("testrunner", "testdata", "concept", "non_executed_tests"),
			expected: filepath.Join("testrunner", "testdata", "expected", "non_executed_tests.json"),
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
			err := os.RemoveAll("outdir")
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

			resultBytes, err := os.ReadFile(filepath.Join("outdir", "results.json"))
			require.NoError(t, err, "failed to read results")

			result := sanitizeResult(string(resultBytes), []string{goExe, currentDir, goRoot})

			expected, err := os.ReadFile(tt.expected)
			require.NoError(t, err, "failed to read expected result file")

			assert.JSONEq(t, string(expected), result)
		})
	}
}

func sanitizeResult(s string, paths []string) string {
	result := s

	for _, p := range pathVariations(paths) {
		result = strings.ReplaceAll(result, p, "PATH_PLACEHOLDER")
	}

	if runtime.GOOS == "windows" {
		result = strings.ReplaceAll(result, `\n.//`, `\n./`)
		result = strings.ReplaceAll(result, `\n.\\`, `\n./`)
		result = strings.ReplaceAll(result, `\n.\`, `\n./`)
	}

	for _, replacement := range regexReplacements {
		result = replacement.regexp.ReplaceAllString(result, replacement.replaceStr)
	}

	return result
}

func pathVariations(paths []string) []string {
	result := []string{}
	for _, p := range paths {
		normalizedPath := filepath.ToSlash(p)
		result = append(result, normalizedPath)

		if runtime.GOOS == "windows" {
			// On windows, the paths that are included in the test results can have
			// various formats. We try to include all variants here so we catch
			// everything when we do the replace later.
			result = append(result, strings.ReplaceAll(normalizedPath, "/", "//"))
			result = append(result, strings.ReplaceAll(normalizedPath, "/", `\`))
			result = append(result, strings.ReplaceAll(normalizedPath, "/", `\\`))
		}
	}

	return result
}
