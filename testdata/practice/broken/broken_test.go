package gigasecond

import (
	"testing"
	"time"
)

// date formats used in test data
const (
	fmtD  = "2006-01-02"
	fmtDt = "2006-01-02T15:04:05"
)

// Write a function AddGigasecond that works with a date
func TestAddGigasecondDate(t *testing.T) {
	var addCases = []struct {
		description string
		in          string
		want        string
	}{
		{
			"date only specification of time",
			"2011-04-25",
			"2043-01-01T01:46:40",
		}, {
			"second test for date only specification of time",
			"1977-06-13",
			"2009-02-19T01:46:40",
		}, {
			"third test for date only specification of time",
			"1959-07-19",
			"1991-03-27T01:46:40",
		},
	}
	for _, tc := range addCases {
		t.Run(tc.description, func(t *testing.T) {
			in, _ := time.Parse(fmtD, tc.in)
			want, _ := time.Parse(fmtDt, tc.want)
			got := AddGigasecond(in)
			if !got.Equal(want) {
				t.Fatalf(`FAIL: %s AddGigasecond(%s) = %s want %s`,
					tc.description, in, got, want,
				)
			}
		})
	}

}

// Write a function AddGigasecond that works with a date + time
func TestAddGigasecondFullTime(t *testing.T) {
	var addCases = []struct {
		description string
		in          string
		want        string
	}{
		{
			"full time specified",
			"2015-01-24T22:00:00",
			"2046-10-02T23:46:40",
		},
		{
			"full time with day roll-over",
			"2015-01-24T23:59:59",
			"2046-10-03T01:46:39",
		},
	}
	for _, tc := range addCases {
		t.Run(tc.description, func(t *testing.T) {
			in, _ := time.Parse(fmtDt, tc.in)
			want, _ := time.Parse(fmtDt, tc.want)
			got := AddGigasecond(in)
			if !got.Equal(want) {
				t.Fatalf(`AddGigasecond(%s) = %s want %s`, in, got, want)
			}
		})
	}
}
