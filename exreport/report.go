package exreport

// Report represents a report of all tests of an exercise.
type Report struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Tests   []Test `json:"tests"`
}
