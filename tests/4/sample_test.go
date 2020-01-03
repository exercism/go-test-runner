package sample

import (
	"testing"

	"github.com/pkg/errors"
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
			if tt.fail {
				t.Error(errors.New("some error"))
			}
		})
	}
}
