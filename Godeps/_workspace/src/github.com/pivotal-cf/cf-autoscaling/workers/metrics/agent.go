package metrics

import (
	"time"


	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-debug/workers/utilities"
	"github.com/evanfarrar/uaa-sso-golang/uaa"
)

type Runner interface {
	Run()
}

type Agent struct {
	killChan       chan bool
	readingsChan   chan<- models.Reading
	timingChan     <-chan time.Time
	serviceBinding models.ServiceBinding
	Interval       time.Duration
	cc             services.CloudControllerInterface
	identifier     string
}

type NullAgent struct{}

func (a NullAgent) Run() {}

func NewAgent(readingsChan chan<- models.Reading,
	serviceBinding models.ServiceBinding,
	interval time.Duration,
	cc services.CloudControllerInterface) Agent {

	identifierCutoff := 5
	if len(serviceBinding.Guid) < 5 {
		identifierCutoff = len(serviceBinding.Guid)
	}
	identifier := serviceBinding.Guid[:identifierCutoff]

	return Agent{
		killChan:       make(chan bool),
		readingsChan:   readingsChan,
		serviceBinding: serviceBinding,
		Interval:       interval,
		timingChan:     time.After(0),
		cc:             cc,
		identifier:     identifier,
	}
}

func (a Agent) Run() {
	for {
		status := a.Work()
		if status != 0 {
			return
		}
	}
}

func (a *Agent) Work() int {
	defer utilities.Recover(a)

	select {
	case <-a.timingChan:
		utilities.Log(a, "Received Timing Event -----> %+v\n", true)
		a.read()
		a.timingChan = time.After(a.Interval)
		return 0
	case <-a.killChan:
		utilities.Log(a, "Received KILL -----> %+v\n", true)
		return 1
	}
}

func (a Agent) Close() {
	a.killChan <- true
	return
}

func (a Agent) read() {
	application, err := a.cc.Stats(a.serviceBinding.AppGuid)
	if err != nil {
		_, isUAAFailure := err.(uaa.Failure)

		switch {
		case err == services.CCErrors.Failure:
			utilities.Log(a, "services.CloudController call failed: %+v\n", err)
			return
		case isUAAFailure:
			utilities.Log(a, "Auth call failed: %+v\n", err)
			return
		default:
			utilities.Log(a, "Unknown error occurred: %+v\n", err)
			return
		}
	}

	reading := models.Reading{
		ServiceBindingGuid:    a.serviceBinding.Guid,
		CreatedAt:             time.Now(),
		CPUUtilization:        application.CPUUtilization,
		ExpectedInstanceCount: application.ExpectedInstanceCount,
		RunningInstanceCount:  application.RunningInstanceCount,
		AppName:               application.Name,
		State:                 application.State,
	}
	utilities.Log(a, "Sending Reading -----> %+v\n", reading)
	a.readingsChan <- reading
}

func (a Agent) Identifier() string {
	return "[Metrics Agent " + a.identifier + "]"
}
