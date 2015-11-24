package util

import "time"

type ClockInterface interface {
	Now() time.Time
}

type Clock struct{}

func NewClock() Clock {
	return Clock{}
}

func (clock Clock) Now() time.Time {
	return time.Now()
}

type TimerInterface interface {
	Tick()
}

type Timer struct {
	duration time.Duration
}

func NewTimer(duration time.Duration) Timer {
	return Timer{
		duration: duration,
	}
}

func (timer Timer) Tick() {
	<-time.After(timer.duration)
}
