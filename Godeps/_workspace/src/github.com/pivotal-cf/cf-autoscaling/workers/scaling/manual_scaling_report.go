package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type ManualScalingReport struct {
	binding models.ServiceBinding
	reading models.Reading
}

func NewManualScalingReport(binding models.ServiceBinding, reading models.Reading) ManualScalingReport {
	return ManualScalingReport{
		binding: binding,
		reading: reading,
	}
}

func (report ManualScalingReport) Binding() models.ServiceBinding {
	return report.binding
}

func (report ManualScalingReport) BuildNotification() services.SpaceNotification {
	context := TemplateContext{
		AppName: report.binding.AppName,
		BindingExpectedInstanceCount: report.binding.ExpectedInstanceCount,
		ReadingExpectedInstanceCount: report.reading.ExpectedInstanceCount,
	}

	renderer := NewTemplateRenderer()

	return services.SpaceNotification{
		KindID:  services.ManualScalingNotification.KindID,
		Subject: services.ManualScalingNotification.Subject,
		Text:    renderer.RenderText("manual_scale_detected.text", context),
		HTML:    renderer.RenderText("manual_scale_detected.html", context),
	}
}

func (report ManualScalingReport) ShouldSend() bool {
	return true
}
