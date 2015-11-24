package scaling

import (
	"fmt"


	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-debug/workers/utilities"
)

type Agent struct {
	ScalingDecision      models.ScalingDecision
	scalingDecisionsRepo models.ScalingDecisionsInterface
	reading              models.Reading
	binding              models.ServiceBinding
	decisions            chan<- models.ScalingDecision
	bindingsRepo         models.ServiceBindingsInterface
}

func NewAgent(reading models.Reading, decisions chan<- models.ScalingDecision, bindingsRepo models.ServiceBindingsInterface, scalingDecisionsRepo models.ScalingDecisionsInterface) Agent {
	decision := models.ScalingDecision{
		ReadingID: reading.ID,
	}

	return Agent{
		reading:              reading,
		decisions:            decisions,
		bindingsRepo:         bindingsRepo,
		ScalingDecision:      decision,
		scalingDecisionsRepo: scalingDecisionsRepo,
	}
}

func (agent *Agent) Run() {
	go func() {
		agent.Decide()
	}()
}

func (agent *Agent) Decide() {
	defer utilities.Recover(agent)

	utilities.Log(agent, "Making scaling decision -----> %+v\n", agent.reading)
	agent.findBinding()
	agent.makeScalingDecision()
	agent.updateServiceBinding()
	agent.scaleApp()
	agent.recordScalingDecision()
	agent.notifyDirector()
}

func (agent *Agent) findBinding() {
	var err error
	agent.binding, err = agent.bindingsRepo.Find(agent.reading.ServiceBindingGuid)
	if err != nil {
		utilities.Log(agent, "Could no find binding -----> %+v\n", agent.reading.ServiceBindingGuid)
	}
}

func (agent Agent) isAboveCPUMaxThreshold() bool {
	return agent.binding.CPUMaxThreshold < agent.reading.CPUUtilization
}

func (agent Agent) isBelowCPUMinThreshold() bool {
	return agent.binding.CPUMinThreshold > agent.reading.CPUUtilization
}

func (agent Agent) canScaleUp() bool {
	return agent.binding.MaxInstances > agent.reading.ExpectedInstanceCount &&
		agent.binding.MaxInstances > agent.reading.RunningInstanceCount
}

func (agent Agent) canScaleDown() bool {
	return agent.binding.MinInstances < agent.reading.ExpectedInstanceCount &&
		agent.binding.MinInstances < agent.reading.RunningInstanceCount &&
		agent.binding.ExpectedInstanceCount > 1
}

func (agent *Agent) isBelowMinInstancesCount() bool {
	return agent.binding.MinInstances > agent.reading.ExpectedInstanceCount &&
		agent.binding.MinInstances > agent.reading.RunningInstanceCount
}

func (agent *Agent) isAboveMaxInstancesCount() bool {
	return agent.binding.MaxInstances < agent.reading.ExpectedInstanceCount &&
		agent.binding.MaxInstances < agent.reading.RunningInstanceCount
}

func (agent Agent) isRunning() bool {
	return agent.reading.State == services.AppRunning
}

func (agent Agent) scalingDisabled() bool {
	return !agent.binding.Enabled
}

func (agent Agent) manualScalingDetected() bool {
	return agent.reading.DoesNotMatchExpectedInstanceCount(agent.binding)
}

func (agent *Agent) makeScalingDecision() {
	var description string
	var factor int
	var decisionType int

	switch {
	case agent.manualScalingDetected():
		description = "Manual Scaling Detected"
		decisionType = models.ManualScaleDetected
	case agent.scalingDisabled():
		description = "Scaling disabled"
		decisionType = models.ScalingDisabled
	case agent.isRunning():
		description, factor, decisionType = agent.makeScalingDecisionForRunningApp()
	default:
		description = "Did not scale"
	}
	agent.ScalingDecision.Description = description
	agent.ScalingDecision.ScalingFactor = factor
	agent.ScalingDecision.DecisionType = decisionType
	agent.ScalingDecision.ServiceBindingGuid = agent.binding.Guid
}

func (agent *Agent) makeScalingDecisionForRunningApp() (string, int, int) {
	var description string
	var factor int
	var decisionType int

	switch {
	case agent.isBelowMinInstancesCount():
		description, factor = agent.scale(1)
		decisionType = models.ScaleUpAppBelowMinAppInstanceCount
	case agent.isAboveMaxInstancesCount():
		description, factor = agent.scale(-1)
		decisionType = models.ScaleDownAppAboveMaxInstanceCount
	case agent.isAboveCPUMaxThreshold() && agent.canScaleUp():
		description, factor = agent.scale(1)
		decisionType = models.ScaleCPUAboveThreshold
	case agent.isAboveCPUMaxThreshold() && !agent.canScaleUp():
		description = fmt.Sprintf("Maximum instance limit of %d reached", agent.binding.MaxInstances)
		decisionType = models.FailedToScaleMaxInstanceCountReached
	case agent.isBelowCPUMinThreshold() && agent.canScaleDown():
		description, factor = agent.scale(-1)
		decisionType = models.ScaleDownBelowCPUUtilizationThreshold
	case agent.isBelowCPUMinThreshold() && !agent.canScaleDown():
		description = fmt.Sprintf("Minimum instance limit of %d reached", agent.binding.MinInstances)
		decisionType = models.NoScaleAtMinimumInstanceCount
	default:
		description = "Did not scale"
		decisionType = models.NoScaleAppWithinThresholds
	}
	return description, factor, decisionType
}

func (agent *Agent) scale(factor int) (string, int) {
	newCount := agent.reading.ExpectedInstanceCount + factor
	description := fmt.Sprintf("Scaled app from %d to %d instances", agent.reading.ExpectedInstanceCount, newCount)
	return description, factor
}

func (agent *Agent) recordScalingDecision() {
	var err error
	agent.ScalingDecision, err = agent.scalingDecisionsRepo.Create(agent.ScalingDecision)
	if err != nil {
		utilities.Log(agent, "Could not create scaling decision -----> %+v\n", agent.ScalingDecision)
	}
}

func (agent *Agent) updateServiceBinding() {
	if agent.ScalingDecision.ScalingFactor != 0 {
		agent.binding.ExpectedInstanceCount = agent.reading.ExpectedInstanceCount + agent.ScalingDecision.ScalingFactor
		_, err := agent.bindingsRepo.Update(agent.binding)
		if err != nil {
			utilities.Log(agent, "Could not update binding -----> %+v\n", agent.binding)
		}
	}
}

func (agent *Agent) scaleApp() {
	if agent.ScalingDecision.ScalingFactor != 0 {
		cc := services.NewCloudControllerClient()
		_, err := cc.Scale(agent.binding.AppGuid, agent.binding.ExpectedInstanceCount)
		if err != nil {
			if err == services.CCErrors.AppQuotaLimitReached {
				agent.ScalingDecision.ScalingFactor = 0
				agent.ScalingDecision.Description = "Failed to scale application due to quota limits"
				agent.ScalingDecision.DecisionType = models.FailedToScaleQuotaDisallows
				agent.binding.ExpectedInstanceCount = agent.reading.RunningInstanceCount
				_, err = agent.bindingsRepo.Update(agent.binding)
				if err != nil {
					utilities.Log(agent, "Could not update binding -----> %+v\n", agent.binding)
				}
				return
			}
			utilities.Log(agent, "Could not scale app with guid -----> %+v\n", agent.binding.AppGuid)
		}
	}
}

func (agent Agent) notifyDirector() {
	if agent.decisions != nil {
		utilities.Log(agent, "Notifying director of completion -----> %+v\n", true)
		agent.decisions <- agent.ScalingDecision
	}
}

func (agent Agent) Identifier() string {
	return "[Scaling Agent]"
}
