package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type MetricsScaleDownReport struct {
	decision models.ScalingDecision
	reading  models.Reading
	binding  models.ServiceBinding
	plan     models.Plan
}

func NewMetricsScaleDownReport(decision models.ScalingDecision, reading models.Reading, binding models.ServiceBinding, plan models.Plan) MetricsScaleDownReport {
	return MetricsScaleDownReport{
		decision: decision,
		reading:  reading,
		binding:  binding,
		plan:     plan,
	}
}

func (report MetricsScaleDownReport) Binding() models.ServiceBinding {
	return report.binding
}

func (report MetricsScaleDownReport) ShouldSend() bool {
	return true
}

func (report MetricsScaleDownReport) BuildNotification() services.SpaceNotification {
	context := TemplateContext{
		AppName:           report.binding.AppName,
		MinInstanceCount:  report.binding.MinInstances,
		MaxInstanceCount:  report.binding.MaxInstances,
		CPUMaxThreshold:   report.binding.CPUMaxThreshold,
		CPUMinThreshold:   report.binding.CPUMinThreshold,
		FromInstanceCount: report.reading.RunningInstanceCount,
		ToInstanceCount:   report.reading.RunningInstanceCount + report.decision.ScalingFactor,
		CPUUtilization:    report.reading.CPUUtilization,
		PlanDuration:      report.plan.FormattedDuration(),
	}

	renderer := NewTemplateRenderer()

	return services.SpaceNotification{
		KindID:  services.ScaleDownNotification.KindID,
		Subject: services.ScaleDownNotification.Subject,
		Text:    renderer.RenderText("scale_down.text", context),
		HTML:    renderer.RenderHTML("scale_down.html", context),
	}
}
