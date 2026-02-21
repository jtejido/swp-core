package clock

import "time"

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

var System Clock = systemClock{}

func Now(c Clock) time.Time {
	if c == nil {
		return System.Now()
	}
	return c.Now()
}

func UnixMilli(c Clock) uint64 {
	return uint64(Now(c).UnixMilli())
}
