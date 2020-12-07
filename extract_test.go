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
