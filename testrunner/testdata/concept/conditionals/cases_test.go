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
