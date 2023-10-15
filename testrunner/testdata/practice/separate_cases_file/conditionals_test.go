package conditionals

import (
	"fmt"
	"testing"
)

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
	fmt.Println(someAssignment)

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
	fmt.Println(someAssignment2)

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
