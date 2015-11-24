package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type OutOfBoundsScalingReport struct {
	decision models.ScalingDecision
	binding  models.ServiceBinding
}

func NewOutOfBoundsScalingReport(decision models.ScalingDecision, binding models.ServiceBinding) OutOfBoundsScalingReport {
	return OutOfBoundsScalingReport{
		decision: decision,
		binding:  binding,
	}
}

func (report OutOfBoundsScalingReport) Binding() models.ServiceBinding {
	return report.binding
}

func (report OutOfBoundsScalingReport) BuildNotification() services.SpaceNotification {
	return services.SpaceNotification{}
}

func (report OutOfBoundsScalingReport) ShouldSend() bool {
	return false
}
