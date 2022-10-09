package externaldeps

import "testing"

func TestIsBigger(t *testing.T) {
	a, b := 1, 2
	if IsBigger(a, b) {
		t.Fatal("a is not bigger than b")
	}
}
