package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type ScheduledScalingReport struct {
	decision models.ScalingDecision
	binding  models.ServiceBinding
}

func NewScheduledScalingReport(decision models.ScalingDecision, binding models.ServiceBinding) ScheduledScalingReport {
	return ScheduledScalingReport{
		decision: decision,
		binding:  binding,
	}
}

func (report ScheduledScalingReport) Binding() models.ServiceBinding {
	return report.binding
}

func (report ScheduledScalingReport) BuildNotification() services.SpaceNotification {
	return services.SpaceNotification{}
}

func (report ScheduledScalingReport) ShouldSend() bool {
	return false
}
