package sample

import "fmt"

// FuncToTest ...
func FuncToTest(someVal string) {
	fmt.Sprintf(someVal)
	fmt.Errorf("some value %s", someVal)
}
