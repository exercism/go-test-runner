package count

import (
	"testing"
)

func TestLocalizedCount(t *testing.T) {
	actual := LocalizedCount()
	if actual != "1,234" {
		t.Fatalf("invalid")
	}
}
