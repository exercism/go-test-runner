package testrunner

import (
	"testing"
)

func TestGetFuncCode(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		testFile string
		code     string
	}{
		{
			name:     "valid call",
			testName: "TestNonSubtest",
			testFile: "testdata/concept/conditionals/conditionals_test.go",
			code:     "func TestNonSubtest(t *testing.T) {\n\t// comments should be included\n\tfmt.Println(\"the whole block\")\n\tfmt.Println(\"should be returned\")\n}",
		}, {
			name:     "missing test",
			testName: "TestNothing",
			testFile: "testdata/concept/conditionals/conditionals_test.go",
			code:     "",
		}, {
			name:     "invalid test file",
			testName: "TestNonSubtest",
			testFile: "testdata/concept/conditionals/conditionals_missing.go",
			code:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code := getFuncCode(tt.testName, tt.testFile); code != tt.code {
				t.Errorf("getFuncCode(%v, %v) = %v; want %v",
					tt.testName, tt.testFile, code, tt.code)
			}
		})
	}
}
