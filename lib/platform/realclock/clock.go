package realclock

import (
	"time"

	"example.com/gotorrent/lib/core/adapter/clock"
)

type RealClock struct{}

var _ clock.Clock = RealClock{}

func (RealClock) Now() time.Time {
	return time.Now()
}

func (RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
