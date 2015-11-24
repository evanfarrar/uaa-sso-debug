package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type QuotaFailureReport struct {
	binding models.ServiceBinding
	reading models.Reading
}

func NewQuotaFailureReport(binding models.ServiceBinding, reading models.Reading) QuotaFailureReport {
	return QuotaFailureReport{
		binding: binding,
		reading: reading,
	}
}

func (report QuotaFailureReport) Binding() models.ServiceBinding {
	return report.binding
}

func (report QuotaFailureReport) ShouldSend() bool {
	return true
}

func (report QuotaFailureReport) BuildNotification() services.SpaceNotification {
	context := TemplateContext{
		AppName:         report.binding.AppName,
		CPUUtilization:  report.reading.CPUUtilization,
		CPUMaxThreshold: report.binding.CPUMaxThreshold,
	}

	renderer := NewTemplateRenderer()

	return services.SpaceNotification{
		KindID:  services.QuotaLimitNotification.KindID,
		Subject: services.QuotaLimitNotification.Subject,
		Text:    renderer.RenderText("quota_failure.text", context),
		HTML:    renderer.RenderText("quota_failure.html", context),
	}
}
