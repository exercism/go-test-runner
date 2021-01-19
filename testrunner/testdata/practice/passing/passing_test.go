package gigasecond

import (
	"fmt"
	"testing"
)

// Trivial passing test example
func TestTrivialPass(t *testing.T) {
	if true != true {
		t.Fatal("This was supposed to be a tautological statement!")
	}
	fmt.Println("sample passing test output")
}
