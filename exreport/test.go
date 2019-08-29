package exreport

// Test represents the result of a test
type Test struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Message  string `json:"message"`
}
