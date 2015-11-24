package models

import (
	"fmt"
	"time"
)

type Plan struct {
	Guid            string
	PollingInterval time.Duration
}

func (plan Plan) FormattedDuration() string {
	duration := plan.PollingInterval
	if duration.Hours() > 1 {
		return fmt.Sprintf("%d hours", int(duration.Hours()))
	}
	if duration.Minutes() > 1 {
		return fmt.Sprintf("%d minutes", int(duration.Minutes()))
	}
	if duration.Seconds() >= 1 {
		return fmt.Sprintf("%d seconds", int(duration.Seconds()))
	}
	if duration.Nanoseconds() > 1 {
		return fmt.Sprintf("%d nanoseconds", int(duration.Nanoseconds()))
	}

	return ""
}
