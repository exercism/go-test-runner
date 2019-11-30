package sample

import (
	"testing"
)

var tests1 = []struct {
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
		fail: false,
	},
}

func TestSample1(t *testing.T) {
	for _, tt := range tests1 {
		if tt.fail {
			t.Fail()
		}
	}
}
