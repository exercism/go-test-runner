// https://exercism.io/tracks/go/exercises/gigasecond
package gigasecond

// import path for the time package from the standard library
import "time"

const Gigasecond = time.Second * 1e9

// Add a Gigasecond (10^9) to the input time
func AddGigasecond(t time.Time) time.Time {
	return t.Add(Gigasecond)
}
