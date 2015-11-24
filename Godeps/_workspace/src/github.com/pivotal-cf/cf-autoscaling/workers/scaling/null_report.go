package scaling

import (

	"github.com/evanfarrar/uaa-sso-debug/services"
)

type NullReport struct{}

func (report NullReport) BuildNotification() services.SpaceNotification {
	return services.SpaceNotification{}
}

func (report NullReport) Binding() models.ServiceBinding {
	return models.ServiceBinding{}
}

func (report NullReport) ShouldSend() bool {
	return false
}
