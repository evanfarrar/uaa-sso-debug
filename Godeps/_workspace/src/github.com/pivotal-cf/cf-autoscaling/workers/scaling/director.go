package scaling

import (
	"github.com/evanfarrar/uaa-sso-debug/exchange"

	"github.com/evanfarrar/uaa-sso-debug/workers/utilities"
)

type Director struct {
	exchange      exchange.Exchange
	readingsSub   exchange.Subscription
	decisions     chan models.ScalingDecision
	kill          chan chan bool
	HandleReading func(models.Reading)
	bindingsRepo  models.ServiceBindingsInterface
	decisionsRepo models.ScalingDecisionsInterface
}

func NewDirector(ex exchange.Exchange, bindingsRepo models.ServiceBindingsInterface, decisionsRepo models.ScalingDecisionsInterface) Director {
	director := Director{
		decisions:     make(chan models.ScalingDecision),
		kill:          make(chan chan bool),
		exchange:      ex,
		readingsSub:   ex.Subscribe("readings:new"),
		bindingsRepo:  bindingsRepo,
		decisionsRepo: decisionsRepo,
	}
	director.HandleReading = director._handleReading
	return director
}

func (d Director) Run() {
	go func() {
		for {
			status := d.Work()
			if status != 0 {
				utilities.Log(d, "Exiting -----> %+v\n", true)
				return
			}
		}
	}()
}

func (d Director) Close() {
	ack := make(chan bool)
	d.kill <- ack
	<-ack
	return
}

func (d Director) Work() int {
	defer utilities.Recover(d)

	select {
	case message := <-d.readingsSub.Channel:
		reading := message.(models.Reading)
		utilities.Log(d, "Received new Reading -----> %+v\n", reading)
		d.HandleReading(reading)
		return 0
	case decision := <-d.decisions:
		utilities.Log(d, "Agent has completed -----> %+v\n", true)
		d.exchange.Publish("scaling-decisions:new", decision)
		return 0
	case ack := <-d.kill:
		utilities.Log(d, "Received KILL -----> %+v\n", true)
		ack <- true
		return 1
	}
}

func (d Director) _handleReading(reading models.Reading) {
	agent := NewAgent(reading, d.decisions, d.bindingsRepo, d.decisionsRepo)
	agent.Run()
}

func (d Director) Identifier() string {
	return "[Scaling Director]"
}
