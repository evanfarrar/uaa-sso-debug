package workers

import (
	"fmt"
	"sort"
	"time"


	"github.com/evanfarrar/uaa-sso-debug/util"
	"github.com/evanfarrar/uaa-sso-debug/workers/utilities"
)

const SchedulingDirectorTimestampKey = "LastExecutionTime"

type SchedulingDirector struct {
	kvRepo               models.KeyValueInterface
	rulesRepo            models.ScheduledRulesInterface
	bindingsRepo         models.ServiceBindingsInterface
	scalingDecisionsRepo models.ScalingDecisionsInterface
	clock                util.ClockInterface
	timer                util.TimerInterface
	die                  bool
}

func NewSchedulingDirector(kvRepo models.KeyValueInterface,
	rulesRepo models.ScheduledRulesInterface,
	bindingsRepo models.ServiceBindingsInterface,
	scalingDecisionsRepo models.ScalingDecisionsInterface,
	clock util.ClockInterface,
	timer util.TimerInterface) SchedulingDirector {

	return SchedulingDirector{
		kvRepo:               kvRepo,
		rulesRepo:            rulesRepo,
		bindingsRepo:         bindingsRepo,
		scalingDecisionsRepo: scalingDecisionsRepo,
		clock:                clock,
		timer:                timer,
	}
}

func (director *SchedulingDirector) Run() {
	utilities.Log(director, "Starting...")
	go func() {
		for {
			director.Work()
			director.timer.Tick()
			if director.die {
				utilities.Log(director, "Dying...")
				return
			}
		}
	}()
}

func (director *SchedulingDirector) Close() {
	director.die = true
	utilities.Log(director, "Received kill order...")
}

func (director *SchedulingDirector) Work() {
	defer utilities.Recover(director)

	lastLookup := director.lastLookupTime()
	lookupTime := director.clock.Now()
	if !lastLookup.IsZero() {
		utilities.Log(director, "Looking up events between %s -> %s...", lastLookup.UTC(), lookupTime.UTC())
		sortableEvents, err := director.rulesRepo.EventsExecutingWithin(lastLookup, lookupTime.Sub(lastLookup))
		if err != nil {
			panic(err)
		}
		director.executeEvents(sortableEvents)
	}
	director.setLookupTime(lookupTime)
}

func (director SchedulingDirector) lastLookupTime() time.Time {
	timestamp, err := director.kvRepo.Get(SchedulingDirectorTimestampKey)
	if err != nil {
		panic(err)
	}

	if timestamp == "" {
		return time.Time{}
	}
	lookupTime, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		panic(err)
	}

	return lookupTime
}

func (director SchedulingDirector) setLookupTime(lookupTime time.Time) {
	err := director.kvRepo.Set(SchedulingDirectorTimestampKey, lookupTime.Format(time.RFC3339Nano))
	if err != nil {
		panic(err)
	}
}

func (director SchedulingDirector) executeEvents(events models.SortableScheduledEvents) {
	sort.Sort(events)
	for _, event := range events.NonOverlappingEvents() {
		utilities.Log(director, "Executing on event: %+v\n", event)
		binding, err := director.bindingsRepo.Find(event.ServiceBindingGuid)
		if err != nil {
			panic(err)
		}

		binding.MinInstances = event.MinInstances
		binding.MaxInstances = event.MaxInstances
		_, err = director.bindingsRepo.Update(binding)
		if err != nil {
			panic(err)
		}

		scalingDescription := fmt.Sprintf("Rule Applied: Scaling Limits set to %d to %d instances", event.MinInstances, event.MaxInstances)
		scalingDecision := models.ScalingDecision{
			Description:        scalingDescription,
			ScalingFactor:      0,
			ServiceBindingGuid: event.ServiceBindingGuid,
		}
		director.scalingDecisionsRepo.Create(scalingDecision)
	}
}

func (director SchedulingDirector) Identifier() string {
	return "[SchedulingDirector]"
}
