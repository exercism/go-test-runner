package testrunner

import (
	"path/filepath"
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
			testFile: filepath.Join("testdata", "concept", "conditionals", "conditionals_test.go"),
			code:     "func TestNonSubtest(t *testing.T) {\n\t// comments should be included\n\tfmt.Println(\"the whole block\")\n\tfmt.Println(\"should be returned\")\n}",
		},
		{
			name:     "invalid test file",
			testName: "TestNonSubtest",
			testFile: filepath.Join("testdata", "concept", "conditionals", "conditionals_missing.go"),
			code:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootLevelTests := FindAllRootLevelTests(tt.testFile)
			rootLevelTestsMap := ConvertToMapByTestName(rootLevelTests)
			if rootLevelTestsMap[tt.testName].code != tt.code {
				t.Errorf("FindAllRootLevelTests for %s did not return correct code, got %v; want %v",
					tt.testFile, rootLevelTestsMap[tt.testName].code, tt.code)
			}
		})
	}
}
