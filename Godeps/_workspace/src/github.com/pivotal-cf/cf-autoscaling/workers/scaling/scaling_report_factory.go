package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type ScalingReport interface {
	BuildNotification() services.SpaceNotification
	Binding() models.ServiceBinding
	ShouldSend() bool
}

type ScalingReportFactory struct {
	decisionsRepo models.ScalingDecisionsInterface
	bindingsRepo  models.ServiceBindingsInterface
	instancesRepo models.ServiceInstancesInterface
	plansRepo     models.PlansInterface
	readingsRepo  models.ReadingsInterface
}

type ScalingReportFactoryInterface interface {
	NewScalingReport(models.ScalingDecision) ScalingReport
}

func NewScalingReportFactory(decisionsRepo models.ScalingDecisionsInterface, bindingsRepo models.ServiceBindingsInterface,
	instancesRepo models.ServiceInstancesInterface, plansRepo models.PlansInterface,
	readingsRepo models.ReadingsInterface) ScalingReportFactory {

	return ScalingReportFactory{
		decisionsRepo: decisionsRepo,
		bindingsRepo:  bindingsRepo,
		instancesRepo: instancesRepo,
		plansRepo:     plansRepo,
		readingsRepo:  readingsRepo,
	}
}

func (factory ScalingReportFactory) NewScalingReport(decision models.ScalingDecision) ScalingReport {
	binding, plan, reading, err := factory.fetch(decision)

	if err != nil {
		return NullReport{}
	}

	switch {
	case factory.isScheduledScaling(decision):
		return NewScheduledScalingReport(decision, binding)
	case factory.isOutOfBoundsScaling(decision):
		return NewOutOfBoundsScalingReport(decision, binding)
	case decision.DecisionType == models.FailedToScaleMaxInstanceCountReached:
		return NewMaxInstanceReachedScalingReport(decision, binding, reading, factory.decisionsRepo)
	case decision.ScalingFactor > 0:
		return NewMetricsScaleUpReport(decision, reading, binding, plan)
	case decision.ScalingFactor < 0:
		return NewMetricsScaleDownReport(decision, reading, binding, plan)
	case decision.DecisionType == models.ManualScaleDetected:
		return NewManualScalingReport(binding, reading)
	case decision.DecisionType == models.FailedToScaleQuotaDisallows:
		return NewQuotaFailureReport(binding, reading)
	}

	return NullReport{}
}

func (factory ScalingReportFactory) fetch(decision models.ScalingDecision) (models.ServiceBinding, models.Plan, models.Reading, error) {
	binding, err := factory.bindingsRepo.Find(decision.ServiceBindingGuid)
	if err != nil {
		return models.ServiceBinding{}, models.Plan{}, models.Reading{}, err
	}

	instance, err := factory.instancesRepo.Find(binding.ServiceInstanceGuid)
	if err != nil {
		return models.ServiceBinding{}, models.Plan{}, models.Reading{}, err
	}

	plan, err := factory.plansRepo.Find(instance.PlanGuid)
	if err != nil {
		return models.ServiceBinding{}, models.Plan{}, models.Reading{}, err
	}

	var reading models.Reading
	if decision.ReadingID != 0 {
		reading, err = factory.readingsRepo.Find(decision.ReadingID)
		if err != nil {
			return models.ServiceBinding{}, models.Plan{}, models.Reading{}, err
		}
	}

	return binding, plan, reading, nil
}

func (factory ScalingReportFactory) isScheduledScaling(decision models.ScalingDecision) bool {
	return decision.IsScheduled()
}

func (factory ScalingReportFactory) isOutOfBoundsScaling(decision models.ScalingDecision) bool {
	return decision.DecisionType == models.ScaleUpAppBelowMinAppInstanceCount || decision.DecisionType == models.ScaleDownAppAboveMaxInstanceCount
}
