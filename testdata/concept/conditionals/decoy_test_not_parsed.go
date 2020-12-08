package conditionals

import (
	"testing"
)

func TestDecoy(t *testing.T) {
	tt := struct {
		name string
		card string
		want int
	}{
		name: "decoy test",
		card: "joker",
		want: -5,
	}
	t.Run(tt.name, func(t *testing.T) {
		if got := ParseCard(tt.card); got != tt.want {
			t.Errorf("TestDecoy(%s) = %d, want %d", tt.card, got, tt.want)
		}
	})
}
