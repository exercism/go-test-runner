package exreport

// Report represents a report of all tests of an exercise.
type Report struct {
	Status string `json:"status"`
	Tests  []Test `json:"tests"`
}
