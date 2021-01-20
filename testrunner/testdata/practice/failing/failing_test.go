package gigasecond

import (
	"fmt"
	"testing"
)

// Trivial failing test example
func TestTrivialFail(t *testing.T) {
	if false != true {
		t.Fatal("Intentional test failure")
	}
	fmt.Println("sample failing test output")
}
