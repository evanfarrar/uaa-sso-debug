package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type MaxInstanceReachedScalingReport struct {
	decision      models.ScalingDecision
	binding       models.ServiceBinding
	reading       models.Reading
	decisionsRepo models.ScalingDecisionsInterface
}

func NewMaxInstanceReachedScalingReport(decision models.ScalingDecision, binding models.ServiceBinding, reading models.Reading, scalingDecisionsRepo models.ScalingDecisionsInterface) MaxInstanceReachedScalingReport {
	return MaxInstanceReachedScalingReport{
		decision:      decision,
		binding:       binding,
		reading:       reading,
		decisionsRepo: scalingDecisionsRepo,
	}
}

func (report MaxInstanceReachedScalingReport) Binding() models.ServiceBinding {
	return report.binding
}

func (report MaxInstanceReachedScalingReport) BuildNotification() services.SpaceNotification {
	context := TemplateContext{
		AppName:           report.binding.AppName,
		MinInstanceCount:  report.binding.MinInstances,
		MaxInstanceCount:  report.binding.MaxInstances,
		CPUMaxThreshold:   report.binding.CPUMaxThreshold,
		CPUMinThreshold:   report.binding.CPUMinThreshold,
		FromInstanceCount: report.reading.RunningInstanceCount,
		ToInstanceCount:   report.reading.RunningInstanceCount + report.decision.ScalingFactor,
		CPUUtilization:    report.reading.CPUUtilization,
	}

	renderer := NewTemplateRenderer()

	return services.SpaceNotification{
		KindID:  services.MaxInstanceCountNotification.KindID,
		Subject: services.MaxInstanceCountNotification.Subject,
		Text:    renderer.RenderText("max_instance_count.text", context),
		HTML:    renderer.RenderHTML("max_instance_count.html", context),
	}

}

func (report MaxInstanceReachedScalingReport) ShouldSend() bool {
	maxInstanceAlreadyNotified, err := report.decisionsRepo.RecentlyNotifiedMaxInstanceReached(report.binding.Guid)
	if err != nil {
		return true
	}

	return !maxInstanceAlreadyNotified
}
