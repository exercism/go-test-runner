package sample

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.False(t, tt.fail)
	}
}
