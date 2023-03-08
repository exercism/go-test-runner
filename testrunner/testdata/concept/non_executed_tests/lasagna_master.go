package lasagna

// PreparationTim estimates the preparation time based on the number of layers and an average time per layer and returns it.
func PreparationTime(layers []string, avgPrepTime int) int {
	if avgPrepTime == 0 {
		avgPrepTime = 2
	}
	return len(layers) * avgPrepTime
}

// Quantities calculates and returns how many noodles and much sauce are needed for the given layers.
func Quantities(layers []string) (noodles int, sauce float64) {
	panic("Please implement")
}

// AddSecretIngredient replaces the secret ingredient of your list with the last ingredient from your friend's list.
func AddSecretIngredient(friendsList, myList []string) {
	panic("Please implement")
}

// ScaleRecipe makes a new slice of float64s from an input slice scaled by a number of portions.
func ScaleRecipe(list []float64, portions int) []float64 {
	panic("Please implement")
}
