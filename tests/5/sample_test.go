package sample

import (
	"testing"
)

var tests = []struct {
	name string
	fail bool
}{
	{
		name: "first test",
		fail: false,
	},
	{
		name: "second test",
		fail: true,
	},
	{
		name: "third test",
		fail: true,
	},
}

func TestSample(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FuncToTest("output from " + tt.name)
			if tt.fail {
				t.Fail()
			}
		})
	}
}
