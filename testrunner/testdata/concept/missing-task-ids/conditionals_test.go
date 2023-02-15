package conditionals

import (
	"fmt"
	"testing"
)

// This test does not have a task ID.
func TestNonSubtest(t *testing.T) {
	// comments should be included
	fmt.Println("the whole block")
	fmt.Println("should be returned")
}

// testRunnerTaskID=1
func TestSimpleSubtest(t *testing.T) {
	myTests := []struct {
		name string
		card string
		want int
	}{
		{
			name: "parse ace",
			card: "ace",
			want: 11,
		},
	}
	for _, tt := range myTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCard(tt.card); got != tt.want {
				t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
			}
		})
	}
}
