package workers

import (
	"time"


	"github.com/evanfarrar/uaa-sso-debug/workers/utilities"
)

type GarbageCollector struct {
	kill             chan chan bool
	timer            <-chan time.Time
	interval         time.Duration
	readings         models.ReadingsInterface
	scalingDecisions models.ScalingDecisionsInterface
	scheduledRules   models.ScheduledRulesInterface
	keyValue         models.KeyValueInterface
}

func NewGarbageCollector(interval time.Duration,
	readings models.ReadingsInterface,
	scalingDecisions models.ScalingDecisionsInterface,
	scheduledRules models.ScheduledRulesInterface,
	keyValue models.KeyValueInterface) GarbageCollector {

	return GarbageCollector{
		kill:             make(chan chan bool),
		timer:            time.After(0),
		interval:         interval,
		readings:         readings,
		scalingDecisions: scalingDecisions,
		scheduledRules:   scheduledRules,
		keyValue:         keyValue,
	}
}

func (gc GarbageCollector) Run() {
	go func() {
		for {
			status := gc.Mux()
			if status != 0 {
				return
			}
		}
	}()
}

func (gc GarbageCollector) Close() {
	ack := make(chan bool)
	gc.kill <- ack
	<-ack
	return
}

func (gc *GarbageCollector) Mux() int {
	defer utilities.Recover(gc)
	select {
	case <-gc.timer:
		utilities.Log(gc, "Pruning Scaling Decisions and Readings tables")
		gc.collect()
		gc.timer = time.After(gc.interval)
		return 0
	case ack := <-gc.kill:
		utilities.Log(gc, "Received Kill")
		ack <- true
		return 1
	}
}

func (gc GarbageCollector) collect() {
	gc.scalingDecisions.DeleteAllBefore(time.Now().Add(-24 * time.Hour))
	gc.readings.DeleteAllBefore(time.Now().Add(-24 * time.Hour))

	lastLookup, err := gc.keyValue.Get(SchedulingDirectorTimestampKey)
	if err != nil {
		utilities.Log(gc, "Failed to fetch previous GC timestamp")
		return
	}

	lastLookupTime, err := time.Parse(time.RFC3339Nano, lastLookup)
	if err == nil {
		gc.scheduledRules.DeleteAllNonRecurringBefore(lastLookupTime)
	}
}

func (gc GarbageCollector) Identifier() string {
	return "[GC]"
}
