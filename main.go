package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/exercism/go-test-runner/testrunner"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: go-test-runner input_dir output_dir")
	}
	input_dir := os.Args[1]
	output_dir := os.Args[2]
	msg, ok := checkArgs(input_dir, output_dir)
	if !ok {
		log.Fatal(msg)
	}

	report := testrunner.Execute(input_dir)
	results := filepath.Join(output_dir, "results.json")
	err := ioutil.WriteFile(results, report, 0644)
	if err != nil {
		log.Fatalf("Failed to write results.json: %s", err)
	}
}

func checkArgs(input_dir string, output_dir string) (string, bool) {
	if _, err := os.Stat(input_dir); os.IsNotExist(err) {
		return fmt.Sprintf("input_dir does not exist: %s", input_dir), false
	}

	if _, err := os.Stat(output_dir); os.IsNotExist(err) {
		log.Printf("output_dir does not exist, creating: %s", output_dir)
		if err := os.Mkdir(output_dir, os.ModeDir|0755); err != nil {
			msg := fmt.Sprintf("output_dir %s does not exist, mkdir failed: %s",
				output_dir, err,
			)
			return msg, false
		}
	}
	return "", true
}
