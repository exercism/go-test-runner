package main

import (
	"strings"
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
			testName: "TestBlackjack",
			codePath: "testdata/concept/conditionals",
			fileName: "testdata/concept/conditionals/conditionals_test.go",
		}, {
			name:     "found subtest",
			testName: "TestBlackjack/blackjack_with_jack_(ace_first)",
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
	tf := "testdata/concept/conditionals/conditionals_test.go"
	tests := []struct {
		name     string
		testName string
		testFile string
		code     string
	}{
		{
			name:     "working regular test",
			testName: "TestNonSubtest",
			testFile: tf,
			code:     "func TestNonSubtest(t *testing.T) {\n\t// comments should be included\n\tfmt.Println(\"the whole block\")\n\tfmt.Println(\"should be returned\")\n}",
		}, {
			name:     "regular test missing subtest",
			testName: "TestNonSubtest/nodice",
			testFile: tf,
			code:     "func TestNonSubtest(t *testing.T) {\n\t// comments should be included\n\tfmt.Println(\"the whole block\")\n\tfmt.Println(\"should be returned\")\n}",
		}, {
			name:     "working subtest",
			testName: "TestParseCard/parse_jack",
			testFile: tf,
			code: `func TestParseCard(t *testing.T) {
	tt := struct {
		name string
		card string
		want int
	}{
	
		name: "parse jack",
		card: "jack",
		want: 10,
	}
	
	if got := ParseCard(tt.card); got != tt.want {
		t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
	}

}`,
		}, {
			name:     "missing / not found subtest",
			testName: "TestParseCard/parse_missing_subtests",
			testFile: "testdata/concept/conditionals/conditionals_test.go",
			code: `func TestParseCard(t *testing.T) {
	tests := []struct {
		name string
		card string
		want int
	}{
		{
			name: "parse two",
			card: "two",
			want: 2,
		},
		{
			name: "parse jack",
			card: "jack",
			want: 10,
		},
		{
			name: "parse king",
			card: "king",
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCard(tt.card); got != tt.want {
				t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
			}
		})
	}
  }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// whitespace / tabs were difficult to match between the test files
			// and the test code / strings... so strip it
			code := extractTestCode(tt.testName, tt.testFile)
			code = strings.Join(strings.Fields(code), " ")
			ttcode := strings.Join(strings.Fields(tt.code), " ")
			if code != ttcode {
				t.Errorf("extractTestCode(%v, %v) = \n%v\n; want \n%v",
					tt.testName, tt.testFile, code, ttcode)
			}
		})
	}
}
