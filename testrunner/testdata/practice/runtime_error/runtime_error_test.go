package gigasecond

import (
	"testing"
	"time"
)

func TestAddGigasecond(t *testing.T) {
	input, _ := time.Parse("2006-01-02", "2011-04-25")
	AddGigasecond(input)
}
