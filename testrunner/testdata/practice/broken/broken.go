// https://exercism.io/tracks/go/exercises/gigasecond
package gigasecond

import "time"

const Gigasecond = time.Second * 1e9

// Add a Gigasecond (10^9) to the input time
func AddGigasecond(t time.Time) time.Time {
	// intentional compilation error
	hmm = nil
	return t.Add(Gigasecond)
}
