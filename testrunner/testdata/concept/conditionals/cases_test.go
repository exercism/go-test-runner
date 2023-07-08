package conditionals

type allergicToInput struct {
	allergen string
	score    uint
}

var allergicToTests = []struct {
	description string
	input       allergicToInput
	expected    bool
}{
	{
		description: "not allergic to anything",
		input: allergicToInput{
			allergen: "eggs",
			score:    0,
		},
		expected: false,
	},
}
var testcases = []struct {
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

type hand struct {
	card1, card2 string
}

var testcases2 = []struct {
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
		hand: hand{
			card1: "ace", card2: "queen"
		},
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
