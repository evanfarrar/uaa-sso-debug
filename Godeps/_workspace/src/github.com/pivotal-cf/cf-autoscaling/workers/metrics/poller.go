package metrics

import (
	"time"


	"github.com/evanfarrar/uaa-sso-debug/workers/utilities"
)

type Bindings []models.ServiceBinding

func (bindings Bindings) Contains(other models.ServiceBinding) bool {
	for _, binding := range bindings {
		if binding == other {
			return true
		}
	}
	return false
}

type PollerInterface interface {
	Run()
	Close()
	Add() <-chan string
	Remove() <-chan string
}

type Poller struct {
	bindingsRepo    models.ServiceBindingsInterface
	Bindings        map[string]models.ServiceBinding
	add             chan string
	remove          chan string
	halt            bool
	pollingInterval time.Duration
}

func NewPoller(bindingsRepo models.ServiceBindingsInterface, pollingInterval time.Duration) *Poller {
	poller := Poller{
		bindingsRepo:    bindingsRepo,
		Bindings:        make(map[string]models.ServiceBinding),
		add:             make(chan string),
		remove:          make(chan string),
		pollingInterval: pollingInterval,
	}

	bindings, err := bindingsRepo.FindAllEnabled()
	if err != nil {
		utilities.Log(poller, "Failed to find enabled bindings: %+v\n", err)
		return &poller
	}

	for _, binding := range bindings {
		poller.Bindings[binding.Guid] = binding
	}

	return &poller
}

func (poller *Poller) Add() <-chan string {
	return poller.add
}

func (poller *Poller) Remove() <-chan string {
	return poller.remove
}

func (poller *Poller) Work() {
	enabledBindings, err := poller.bindingsRepo.FindAllEnabled()
	if err != nil {
		utilities.Log(poller, "Failed to find enabled bindings: %+v\n", err)
		return
	}

	for _, binding := range enabledBindings {
		if _, ok := poller.Bindings[binding.Guid]; !ok {
			poller.Bindings[binding.Guid] = binding
			poller.add <- binding.Guid
		}
	}

	for guid, binding := range poller.Bindings {
		if !Bindings(enabledBindings).Contains(binding) {
			delete(poller.Bindings, guid)
			poller.remove <- guid
		}
	}
}

func (poller *Poller) Run() {
	go func() {
		for {
			if poller.halt {
				return
			}
			poller.Work()
			<-time.After(poller.pollingInterval)
		}
	}()
}

func (poller *Poller) Close() {
	poller.halt = true
}

func (poller Poller) Identifier() string {
	return "[Metrics Poller]"
}
