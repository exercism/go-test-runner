package conditionals

import (
	"fmt"
	"testing"
)

// testRunnerTaskID=no-ID
func TestNonSubtest(t *testing.T) {
	// comments should be included
	fmt.Println("the whole block")
	fmt.Println("should be returned")
}

// testRunnerTaskID=2
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

// testRunnerTaskID=1
func TestParseCard(t *testing.T) {
	tests := []struct {
		name string
		card string
		want int
	}{
		{
			name: "parse two",
			card: "two",
			want: 2,
		},
		{
			name: "parse jack",
			card: "jack",
			want: 10,
		},
		{
			name: "parse king",
			card: "king",
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCard(tt.card); got != tt.want {
				t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
			}
		})
	}
}

// testRunnerTaskID=3
func TestBlackjack(t *testing.T) {
	someAssignment := "test"
	fmt.Println(someAssignment)

	type hand struct {
		card1, card2 string
	}
	tests := []struct {
		name string
		hand hand
		want bool
	}{
		{
			name: "blackjack with ten (ace first)",
			hand: hand{card1: "ace", card2: "ten"},
			want: true,
		},
		{
			name: "blackjack with jack (ace first)",
			hand: hand{card1: "ace", card2: "jack"},
			want: true,
		},
		{
			name: "blackjack with queen (ace first)",
			hand: hand{card1: "ace", card2: "queen"},
			want: true,
		},
		{
			name: "blackjack with king (ace first)",
			hand: hand{card1: "ace", card2: "king"},
			want: true,
		},
		{
			name: "no blackjack with eight and five",
			hand: hand{card2: "eight", card1: "five"},
			want: false,
		},
	}

	_ = "literally anything"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBlackjack(tt.hand.card1, tt.hand.card2)
			if got != tt.want {
				t.Errorf("IsBlackjack(%s, %s) = %t, want %t", tt.hand.card1, tt.hand.card2, got, tt.want)
			}
		})
	}

	// Additional statements should be included
	fmt.Println("the whole block")
	fmt.Println("should be returned")
}
