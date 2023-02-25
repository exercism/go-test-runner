package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	tests := []struct {
		inputDir string
		expected string
	}{
		{
			inputDir: "./testrunner/testdata/practice/broken",
			expected: "./testrunner/testdata/expected/practice_broken.json",
		},
	}

	goExe, err := exec.LookPath("go")
	require.NoError(t, err, "failed to find go executable")

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
			require.NoErrorf(t, err, "failed to execute test runner: %s, %s", stdout.String(), stderr.String())
			// require.Errorf(t, err, "failed to execute test runner: %s, %s", stdout.String(), stderr.String())

			result, err := os.ReadFile("./outdir/results.json")
			require.NoError(t, err, "failed to read results")

			sanitizedResult := strings.ReplaceAll(string(result), goExe, "PATH_PLACEHOLDER")

			expected, err := os.ReadFile(tt.expected)
			require.NoError(t, err, "failed to read expected result file")

			assert.JSONEq(t, string(expected), sanitizedResult)
		})
	}
}
