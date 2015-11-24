package metrics

import (
	"github.com/evanfarrar/uaa-sso-debug/exchange"

	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-debug/workers/utilities"
)

type Director struct {
	killChan             chan bool
	agents               map[string]Agent
	readingsChannel      chan models.Reading
	exchange             exchange.Exchange
	newBindingSub        exchange.Subscription
	deleteBindingSub     exchange.Subscription
	serviceBindingsRepo  models.ServiceBindingsInterface
	serviceInstancesRepo models.ServiceInstancesInterface
	plansRepo            models.Plans
	readingsRepo         models.ReadingsInterface
	poller               PollerInterface
}

func NewDirector(serviceBindingsRepo models.ServiceBindingsInterface, serviceInstancesRepo models.ServiceInstancesInterface,
	readingsRepo models.ReadingsInterface, ex exchange.Exchange, poller PollerInterface) Director {

	return Director{
		killChan:             make(chan bool),
		agents:               make(map[string]Agent),
		readingsChannel:      make(chan models.Reading),
		exchange:             ex,
		newBindingSub:        ex.Subscribe("bindings:new"),
		deleteBindingSub:     ex.Subscribe("bindings:delete"),
		serviceBindingsRepo:  serviceBindingsRepo,
		serviceInstancesRepo: serviceInstancesRepo,
		readingsRepo:         readingsRepo,
		plansRepo:            models.Plans{},
		poller:               poller,
	}
}

func (d Director) AgentsCount() int {
	utilities.Log(d, "AGENTS COUNT -----> %+v\n", d.agents)
	return len(d.agents)
}

func (d Director) Run() {
	d.poller.Run()
	go func() {
		d.startAgents()
		for {
			status := d.Listen()
			if status != 0 {
				utilities.Log(d, "Exiting -----> %+v\n", true)
				return
			}
		}
	}()
}

func (d Director) Close() {
	utilities.Log(d, "Received Close -----> %+v\n", true)
	d.poller.Close()
	d.killChan <- true
	return
}

func (d Director) Listen() int {
	defer utilities.Recover(d)
	select {
	case reading := <-d.readingsChannel:
		d.HandleReading(reading)
		return 0
	case message := <-d.poller.Add():
		d.handleNewBinding(message)
		return 0
	case message := <-d.poller.Remove():
		d.handleDeleteBinding(message)
		return 0
	case <-d.killChan:
		d.closeAgents()
		return 1
	}
}

func (d Director) NewAgentFor(sb models.ServiceBinding) Runner {
	if _, ok := d.agents[sb.Guid]; ok {
		return NullAgent{}
	} else {
		si, err := d.serviceInstancesRepo.Find(sb.ServiceInstanceGuid)
		if err != nil {
			utilities.Log(d, "Service instance not found when creating a new agent: %+v\n", err)
			return NullAgent{}
		}

		plan, err := d.plansRepo.Find(si.PlanGuid)
		if err != nil {
			utilities.Log(d, "Plan not found when creating a new agent: %+v\n", err)
			return NullAgent{}
		}

		agent := NewAgent(d.readingsChannel, sb, plan.PollingInterval, services.NewCloudControllerClient())
		d.agents[sb.Guid] = agent
		return agent
	}
}

func (d Director) startAgents() {
	bindings, _ := d.serviceBindingsRepo.FindAllEnabled()
	for _, binding := range bindings {
		utilities.Log(d, "Creating agent for binding -----> %+v\n", binding)
		agent := d.NewAgentFor(binding)
		go agent.Run()
	}
}

func (d Director) handleNewBinding(bindingGuid string) {
	utilities.Log(d, "Received new binding -----> %+v\n", bindingGuid)
	binding, err := d.serviceBindingsRepo.Find(bindingGuid)
	if err != nil {
		utilities.Log(d, "Failed to find service binding -----> %+v\n", bindingGuid)
	}

	agent := d.NewAgentFor(binding)
	utilities.Log(d, "Spawning new AGENT -----> %+v\n", agent)
	go agent.Run()
}

func (d Director) handleDeleteBinding(bindingGuid string) {
	utilities.Log(d, "Received delete binding -----> %+v\n", bindingGuid)
	if agent, ok := d.agents[bindingGuid]; ok {
		d.closeAgent(agent, bindingGuid)
	}
}

func (d Director) closeAgents() {
	for bindingGuid, agent := range d.agents {
		d.closeAgent(agent, bindingGuid)
	}
}

func (d Director) closeAgent(agent Agent, bindingGuid string) {
	agent.Close()
	delete(d.agents, bindingGuid)
}

func (d Director) HandleReading(reading models.Reading) {
	utilities.Log(d, "Received new Reading -----> %+v\n", reading)
	binding, err := d.serviceBindingsRepo.Find(reading.ServiceBindingGuid)
	if err != nil {
		utilities.Log(d, "Failed to find service binding for reading %+v\n", reading)
		return
	}

	if binding.AppName != reading.AppName {
		binding.AppName = reading.AppName
		_, err = d.serviceBindingsRepo.Update(binding)
		if err != nil {
			utilities.Log(d, "Failed to update service binding for reading %+v\n", reading)
			return
		}
	}

	reading, err = d.readingsRepo.Create(reading)
	if err != nil {
		utilities.Log(d, "Failed to create reading %+v\n", reading)
		return
	}

	if reading.DoesNotMatchExpectedInstanceCount(binding) {
		binding.Enabled = false
		d.serviceBindingsRepo.Update(binding)
	}

	d.exchange.Publish("readings:new", reading)
}

func (d Director) Identifier() string {
	return "[Metrics Director]"
}
