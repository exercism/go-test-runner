package main

import (
	"testing"
)

func TestSplitTestName(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		mainTest string
		subTest  string
	}{
		{
			name:     "no subtest",
			testName: "TestParseCard",
			mainTest: "TestParseCard",
			subTest:  "",
		}, {
			name:     "has subtest",
			testName: "TestParseCard/parse_four",
			mainTest: "TestParseCard",
			subTest:  "parse_four",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tn, subtn := splitTestName(tt.testName); tn != tt.mainTest || subtn != tt.subTest {
				t.Errorf("splitTestName(%v) = %v, %v; want %v, %v",
					tt.name, tn, subtn, tt.mainTest, tt.subTest)
			}
		})
	}
}

func TestFindTestFile(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		codePath string
		fileName string
	}{
		{
			name:     "found test",
			testName: "TestFirstTurn",
			codePath: "testdata/concept/conditionals",
			fileName: "testdata/concept/conditionals/conditionals_test.go",
		}, {
			name:     "found subtest",
			testName: "TestFirstTurn/pair_of_jacks",
			codePath: "testdata/concept/conditionals",
			fileName: "testdata/concept/conditionals/conditionals_test.go",
		}, {
			name:     "missing test",
			testName: "TestMissing",
			codePath: "",
			fileName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tf := findTestFile(tt.testName, tt.codePath); tf != tt.fileName {
				t.Errorf("findTestFile(%v, %v) = %v; want %v",
					tt.testName, tt.codePath, tf, tt.fileName)
			}
		})
	}
}

func TestExtractTestCode(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		testFile string
		code     string
	}{
		{
			name:     "found test",
			testName: "TestNonSubtest",
			testFile: "testdata/concept/conditionals/conditionals_test.go",
			code:     "func TestNonSubtest(t *testing.T) {\n\t// comments should be included\n\tfmt.Println(\"the whole block\")\n\tfmt.Println(\"should be returned\")\n}",
		}, /* [TODO] - test a reasonable subtest{
			name:     "found subtest",
			testName: "TestFirstTurn/pair_of_jacks",
			testFile: "testdata/concept/conditionals/conditionals_test.go",
			code:     "pair_of_jacks",
		}, */
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code := extractTestCode(tt.testName, tt.testFile); code != tt.code {
				t.Errorf("extractTestCode(%v, %v) = %v; want %v",
					tt.testName, tt.testFile, code, tt.code)
			}
		})
	}
}
