package exreport

// Test represents the result of a test
type Test struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
