package gigasecond

import (
	"fmt"
	"testing"
)

// Trivial passing test example 1
func TestTrivialPass1(t *testing.T) {
	t.Run("subtest 1.1", func(t *testing.T) {
		if true != true {
			t.Fatal("Should never happen!")
		}
		fmt.Println("sample passing subtest output 1.1")
	})

	t.Run("subtest 1.2", func(t *testing.T) {
		if true != true {
			t.Fatal("Should never happen!")
		}
		fmt.Println("sample passing subtest output 1.2")
	})
}

// Trivial passing test example 2
func TestTrivialPass2(t *testing.T) {
	if true != true {
		t.Fatal("Should never happen!")
	}
	fmt.Println("sample passing test output 2")
}


// Tests skip test
func TestSkip1(t *testing.T) {
	t.Skip("skipped test")
}
