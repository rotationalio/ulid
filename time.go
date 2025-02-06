package ulid

import "time"

// maxTime is the maximum Unix time in milliseconds that can be
// represented in a ULID.
var maxTime = ULID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}.Time()

// MaxTime returns the maximum Unix time in milliseconds that
// can be encoded in a ULID.
func MaxTime() uint64 { return maxTime }

// Now is a convenience function that returns the current
// UTC time in Unix milliseconds. Equivalent to:
//
//	Timestamp(time.Now().UTC())
func Now() uint64 { return Timestamp(time.Now().UTC()) }

// Timestamp converts a time.Time to Unix milliseconds.
//
// Because of the way ULID stores time, times from the year
// 10889 produces undefined results.
func Timestamp(t time.Time) uint64 {
	return uint64(t.Unix())*1000 +
		uint64(t.Nanosecond()/int(time.Millisecond))
}

// Time converts Unix milliseconds in the format
// returned by the Timestamp function to a time.Time.
func Time(ms uint64) time.Time {
	s := int64(ms / 1e3)
	ns := int64((ms % 1e3) * 1e6)
	return time.Unix(s, ns)
}
