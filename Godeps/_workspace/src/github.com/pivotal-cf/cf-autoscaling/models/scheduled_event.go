package models

import (
	"fmt"
	"sort"
	"time"
)

type ScheduledEvent struct {
	ExecutesAt         time.Time
	ServiceBindingGuid string
	MinInstances       int
	MaxInstances       int
	Recurrs            bool
}

type SortableScheduledEvents []ScheduledEvent

func (sortable SortableScheduledEvents) Len() int {
	return len(sortable)
}

func (sortable SortableScheduledEvents) Less(i, j int) bool {
	return sortable[i].ExecutesAt.Before(sortable[j].ExecutesAt)
}

func (sortable SortableScheduledEvents) Swap(i, j int) {
	sortable[j], sortable[i] = sortable[i], sortable[j]
}

func (sortable SortableScheduledEvents) NonOverlappingEvents() SortableScheduledEvents {
	eventsMap := make(map[string]ScheduledEvent)
	for _, event := range sortable {
		key := fmt.Sprintf("%d %s", event.ExecutesAt.Unix(), event.ServiceBindingGuid)
		_, ok := eventsMap[key]
		if !ok || (ok && !event.Recurrs) {
			eventsMap[key] = event
		}
	}
	events := make(SortableScheduledEvents, 0)
	for _, event := range eventsMap {
		events = append(events, event)
	}
	sort.Sort(events)
	return events
}
