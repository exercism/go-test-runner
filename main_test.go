package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func ExampleMain() {
	os.Args = []string{
		"path",
		filepath.Join("testrunner", "testdata", "practice", "passing"),
		"outdir",
	}
	main()
	// Output:
}

func TestMain_fail(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		exitMsg string
	}{
		{
			name: "incorrect usage",
			args: []string{"progpath", "only_one_arg"},
		},
		{
			name: "missing input_dir",
			args: []string{"progpath", "bad_input_dir", "noop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// technique from: https://talks.golang.org/2014/testing.slide#23
			// base case - if the env var is set, just run main()
			if os.Getenv("BE_MAIN") == "1" {
				os.Args = tt.args
				main()
				return
			}

			// exec the test from itself, so we can safely inspect the fatal exit
			tn := fmt.Sprintf("-test.run=TestMain_fail/%s", tt.name)
			cmd := exec.Command(os.Args[0], tn)
			// set the base case env var to avoid infinite recursion
			cmd.Env = append(os.Environ(), "BE_MAIN=1")
			err := cmd.Run()
			if e, ok := err.(*exec.ExitError); ok || !e.Success() {
				return
			}
			t.Fatalf("main() test ran with err %v, want exit status 1", err)
		})
	}
}

func TestCheckArgs(t *testing.T) {
	tests := []struct {
		name       string
		input_dir  string
		output_dir string
		ok         bool
		msg        string
	}{
		{
			name:       "missing input_dir",
			input_dir:  "testrunner",
			output_dir: "testrunner",
			ok:         true,
			msg:        "",
		},
		{
			name:       "missing input_dir",
			input_dir:  "bad_input_dir",
			output_dir: "noop",
			ok:         false,
			msg:        "input_dir does not exist: bad_input_dir",
		},
		{
			name:       "broken output_dir",
			input_dir:  "testrunner",
			output_dir: filepath.Join("/", "tmp", "broken", "rmme"),
			ok:         false,
			msg:        "output_dir " + filepath.Join("/", "tmp", "broken", "rmme") + " does not exist, mkdir failed:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, ok := checkArgs(tt.input_dir, tt.output_dir)
			if ok != tt.ok || !strings.HasPrefix(msg, tt.msg) {
				t.Fatalf("checkArgs(%s, %s) = %s, %t, want %s, HasPrefix(%t)",
					tt.input_dir, tt.output_dir, msg, ok, tt.msg, tt.ok,
				)
			}
		})
	}
}
