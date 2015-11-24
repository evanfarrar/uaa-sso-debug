package models

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

var NoFutureScheduledEventError = errors.New("No future scheduled event")

type ScheduledRule struct {
	ID                 int       `db:"id"                   json:"id"`
	ExecutesAt         time.Time `db:"executes_at"          json:"executes_at"`
	MinInstances       int       `db:"min_instances"        json:"min_instances"`
	MaxInstances       int       `db:"max_instances"        json:"max_instances"`
	ServiceBindingGuid string    `db:"service_binding_guid" json:"service_binding_guid"`
	CreatedAt          time.Time `db:"created_at"           json:"-"`
	Recurrence         int       `db:"recurrence"           json:"recurrence"`
	Enabled            bool      `db:"enabled"              json:"enabled"`
}

func (rule ScheduledRule) Overlaps(other ScheduledRule) bool {
	if rule.Recurrence&other.Recurrence != 0 &&
		rule.ExecutesAt.Format(time.Kitchen) == other.ExecutesAt.Format(time.Kitchen) {
		return true
	}
	return false
}

func (rule ScheduledRule) NextEvent() (ScheduledEvent, error) {
	event := ScheduledEvent{}
	if rule.IsRecurring() {
		oneWeek := 7 * 24 * time.Hour
		events := rule.FutureEvents(rule.eventHorizon(), oneWeek)
		sort.Sort(events)
		return events[0], nil
	} else {
		if rule.InFuture() {
			event.ExecutesAt = rule.ExecutesAt
		} else {
			return event, NoFutureScheduledEventError
		}
	}
	return event, nil
}

func (rule ScheduledRule) IsNew() bool {
	return rule.ID == 0
}

func (rule ScheduledRule) InFuture() bool {
	return rule.ExecutesAt.After(time.Now())
}

func (rule ScheduledRule) IsRecurring() bool {
	return rule.Recurrence != 0
}

func (rule ScheduledRule) IsDisabled() bool {
	return !rule.Enabled
}

func (rule ScheduledRule) FutureEvents(horizon time.Time, duration time.Duration) SortableScheduledEvents {
	events := make(SortableScheduledEvents, 0)
	if rule.IsRecurring() {
		if rule.between(time.Time{}, horizon.Add(duration)) {
			for _, weekday := range rule.recurringWeekdays() {
				executesAt := rule.findNextExecutesAtOnWeekday(weekday, horizon)

				if (executesAt.Equal(horizon) || executesAt.After(horizon)) && executesAt.Before(horizon.Add(duration+1*time.Nanosecond)) {
					events = append(events, rule.ScheduledEventWithExecutesAt(executesAt))
				}
			}
		}
	} else {
		if rule.between(horizon, horizon.Add(duration)) {
			events = append(events, rule.ScheduledEventWithExecutesAt(rule.ExecutesAt))
		}
	}
	return events
}

func (rule ScheduledRule) ScheduledEventWithExecutesAt(executesAt time.Time) ScheduledEvent {
	return ScheduledEvent{
		ExecutesAt:         executesAt,
		MinInstances:       rule.MinInstances,
		MaxInstances:       rule.MaxInstances,
		ServiceBindingGuid: rule.ServiceBindingGuid,
		Recurrs:            rule.IsRecurring(),
	}
}

func (rule ScheduledRule) findNextExecutesAtOnWeekday(weekday int, horizon time.Time) time.Time {
	horizonWeekday := horizon.Weekday()
	diff := (weekday - int(horizonWeekday) + 7) % 7
	diffDuration := time.Duration(int64(diff) * int64(24) * int64(time.Hour))
	futureTime := horizon.Add(diffDuration)

	executesAt := time.Date(
		futureTime.Year(),
		futureTime.Month(),
		futureTime.Day(),
		rule.ExecutesAt.Hour(),
		rule.ExecutesAt.Minute(),
		rule.ExecutesAt.Second(),
		rule.ExecutesAt.Nanosecond(),
		rule.ExecutesAt.Location())
	if executesAt.Before(horizon) || executesAt.Equal(horizon) {
		executesAt = executesAt.Add(7 * 24 * time.Hour)
	}
	return executesAt
}

func (rule ScheduledRule) between(start, end time.Time) bool {
	return rule.ExecutesAt.After(start) && rule.ExecutesAt.Before(end)
}

func (rule ScheduledRule) eventHorizon() time.Time {
	if rule.InFuture() {
		return rule.ExecutesAt
	} else {
		return time.Now().UTC()
	}
}

func (rule ScheduledRule) recurringWeekdays() []int {
	binaryRepresentation := fmt.Sprintf("%07s", strconv.FormatInt(int64(rule.Recurrence), 2))
	weekdays := make([]int, 0)
	for index, i := range strings.Split(binaryRepresentation, "") {
		set, _ := strconv.ParseInt(i, 10, 0)
		if set == 1 {
			weekdays = append(weekdays, int(set)*index)
		}
	}
	return weekdays
}
