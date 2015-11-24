package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type MetricsScaleUpReport struct {
	decision models.ScalingDecision
	reading  models.Reading
	binding  models.ServiceBinding
	plan     models.Plan
}

func NewMetricsScaleUpReport(decision models.ScalingDecision, reading models.Reading, binding models.ServiceBinding, plan models.Plan) MetricsScaleUpReport {
	return MetricsScaleUpReport{
		decision: decision,
		reading:  reading,
		binding:  binding,
		plan:     plan,
	}
}

func (report MetricsScaleUpReport) Binding() models.ServiceBinding {
	return report.binding
}

func (report MetricsScaleUpReport) ShouldSend() bool {
	return true
}

func (report MetricsScaleUpReport) BuildNotification() services.SpaceNotification {
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
		KindID:  services.ScaleUpNotification.KindID,
		Subject: services.ScaleUpNotification.Subject,
		Text:    renderer.RenderText("scale_up.text", context),
		HTML:    renderer.RenderHTML("scale_up.html", context),
	}
}
