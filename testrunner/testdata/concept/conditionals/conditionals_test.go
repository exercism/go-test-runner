package conditionals

import (
	"fmt"
	"testing"
)

func TestNonSubtest(t *testing.T) {
	// comments should be included
	fmt.Println("the whole block")
	fmt.Println("should be returned")
}

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

func TestSimpleSubtest_NoFieldName(t *testing.T) {
	myTests := []struct {
		name string
		card string
		want int
	}{
		{
			"parse ace",
			"ace",
			11,
		},
		{
			"parse two",
			"two",
			2,
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
func TestParseCard_Separate(t *testing.T) {
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCard(tt.card); got != tt.want {
				t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
			}
		})
	}
}

func TestBlackjack_Separate(t *testing.T) {
	someAssignment := "test"
	fmt.Println(someAssignment)

	_ = "literally anything"

	for _, tt := range testcases2 {
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
func TestSubtest_MultiAssignStmt(t *testing.T) {
	someAssignment := "test"
	
	myTests := []struct {
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

	someAssignment2 := "test2"

	for _, tt := range myTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCard(tt.card); got != tt.want {
				t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
			}
		})
	}
	
	// Additional statements should be included
	fmt.Println("the whole block")
	fmt.Println("should be returned")
}
