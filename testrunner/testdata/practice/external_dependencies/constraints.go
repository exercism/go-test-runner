package externaldeps

import (
	"golang.org/x/exp/constraints"
)

func IsBigger[T constraints.Ordered](a T, b T) bool {
	return a > b
}
