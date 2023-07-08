package testrunner

import (
	"path/filepath"
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
		codePath string
		fileName string
	}{
		{
			name:     "found single test file",
			codePath: filepath.Join("testdata", "practice", "passing"),
			fileName: filepath.Join("testdata", "practice", "passing", "passing_test.go"),
		},
		{
			name:     "found correct test file if there are two",
			codePath: filepath.Join("testdata", "concept", "conditionals"),
			fileName: filepath.Join("testdata", "concept", "conditionals", "conditionals_test.go"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tf := FindTestFile(tt.codePath); tf != tt.fileName {
				t.Errorf("findTestFile(%v) = %v; want %v",
					tt.codePath, tf, tt.fileName)
			}
		})
	}
}

func TestExtractTestCode(t *testing.T) {
	tf := filepath.Join("testdata", "concept", "conditionals", "conditionals_test.go")
	rootLevelTests := FindAllRootLevelTests(tf)
	rootLevelTestsMap := ConvertToMapByTestName(rootLevelTests)
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
			name:     "working simple subtest with different name for test data variable",
			testName: "TestSimpleSubtest/parse_ace",
			testFile: tf,
			code: `func TestSimpleSubtest(t *testing.T) {
	tt := struct {
		name string
		card string
		want int
	}{
		name: "parse ace",
		card: "ace",
		want: 11,
	}
	
	if got := ParseCard(tt.card); got != tt.want {
		t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
	}

}`,
		}, {
			name:     "working simple subtest with no field name in test data variable",
			testName: "TestSimpleSubtest_NoFieldName/parse_ace",
			testFile: tf,
			code: `func TestSimpleSubtest_NoFieldName(t *testing.T) {
	tt := struct {
		name string
		card string
		want int
	}{
		"parse ace",
		"ace",
		11,
	}
	
	if got := ParseCard(tt.card); got != tt.want {
		t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
	}

}`,
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
			name:     "subtest with additional code above and below test data, and multiple statements inside Run()",
			testName: "TestBlackjack/blackjack_with_ten_(ace_first)",
			testFile: tf,
			code: `func TestBlackjack(t *testing.T) {
	someAssignment := "test"
	fmt.Println(someAssignment)

	type hand struct {
		card1, card2 string
	}
	tt := struct {
		name string
		hand hand
		want bool
	}{
		name: "blackjack with ten (ace first)",
		hand: hand{card1: "ace", card2: "ten"},
		want: true,
	}

	_ = "literally anything"
	
	got := IsBlackjack(tt.hand.card1, tt.hand.card2)
	if got != tt.want {
		t.Errorf("IsBlackjack(%s, %s) = %t, want %t", tt.hand.card1, tt.hand.card2, got, tt.want)
	}

	// Additional statements should be included
	fmt.Println("the whole block")
	fmt.Println("should be returned")
}`,
		}, {
			name:     "missing / not found subtest",
			testName: "TestParseCard/parse_missing_subtests",
			testFile: filepath.Join("testdata", "concept", "conditionals", "conditionals_test.go"),
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
	tests = append(tests, testsDataSeparate...)
	tests = append(tests, testsMultiAssignStmt...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := ExtractTestCodeAndTaskID(rootLevelTestsMap, tt.testName)

			actualLines := strings.Split(code, "\n")
			expectedLines := strings.Split(tt.code, "\n")

			if len(actualLines) != len(expectedLines) {
				t.Errorf("ExtractTestCode(%v, %v)\n has %v lines\n; want %v lines",
					tt.testName, tt.testFile, len(actualLines), len(expectedLines))
			}

			for i, actual := range actualLines {
				expected := expectedLines[i]
				// whitespace / tabs were difficult to match between the test files
				// and the test code / strings... so strip them
				actual = strings.Join(strings.Fields(actual), " ")
				expected = strings.Join(strings.Fields(expected), " ")

				if actual != expected {
					t.Errorf("ExtractTestCode(%v, %v) = \n%v\n; want\n%v\n"+
						"; differ on line: %v\n; have: `%v`\n; want: `%v`",
						tt.testName, tt.testFile, code, tt.code,
						i+1, actualLines[i], expectedLines[i])
					break
				}
			}
		})
	}
}

var testsDataSeparate = []struct {
	name     string
	testName string
	testFile string
	code     string
}{
	{
		name:     "working subtest with separate test data",
		testName: "TestParseCard_Separate/parse_jack",
		testFile: filepath.Join("testdata", "concept", "conditionals", "conditionals_test.go"),
		code: `func TestParseCard_Separate(t *testing.T) {
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
		name:     "missing / not found subtest with separate test data",
		testName: "TestParseCard_Separate/parse_missing_subtests",
		testFile: filepath.Join("testdata", "concept", "conditionals", "conditionals_test.go"),
		code: `func TestParseCard_Separate(t *testing.T) {
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCard(tt.card); got != tt.want {
				t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
			}
		})
	}
}`,
	}, {
		name:     "multiple statements with separate test data",
		testName: "TestBlackjack_Separate/blackjack_with_ten_(ace_first)",
		testFile: filepath.Join("testdata", "concept", "conditionals", "conditionals_test.go"),
		code: `func TestBlackjack_Separate(t *testing.T) {
	tt := struct {
		name string
		hand hand
		want bool
	}{
		name: "blackjack with ten (ace first)",
		hand: hand{card1: "ace", card2: "ten"},
		want: true,
	}
	someAssignment := "test"
	fmt.Println(someAssignment)

	_ = "literally anything"

	got := IsBlackjack(tt.hand.card1, tt.hand.card2)
	if got != tt.want {
		t.Errorf("IsBlackjack(%s, %s) = %t, want %t", tt.hand.card1, tt.hand.card2, got, tt.want)
	}

	// Additional statements should be included
	fmt.Println("the whole block")
	fmt.Println("should be returned")
}`,
	},
}
var testsMultiAssignStmt = []struct {
	name     string
	testName string
	testFile string
	code     string
}{
	{
		name:     "subtest with arbitrary test data variable name, additional assign statements above and below test data",
		testName: "TestSubtest_MultiAssignStmt/parse_king",
		testFile: filepath.Join("testdata", "concept", "conditionals", "conditionals_test.go"),
		code: `func TestSubtest_MultiAssignStmt(t *testing.T) {
	someAssignment := "test"

	tt := struct {
		name string
		card string
		want int
	}{
		name: "parse king",
		card: "king",
		want: 10,
	}

	someAssignment2 := "test2"

	if got := ParseCard(tt.card); got != tt.want {
		t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
	}

	// Additional statements should be included
	fmt.Println("the whole block")
	fmt.Println("should be returned")
}`,
	},
}
