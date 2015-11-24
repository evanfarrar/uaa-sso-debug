package validators

import "github.com/evanfarrar/uaa-sso-debug/models"

type ServiceBindingValidator struct {
	Errors []string
}

func NewServiceBindingValidator() ServiceBindingValidator {
	return ServiceBindingValidator{}
}

func (s *ServiceBindingValidator) Validate(sb models.ServiceBinding) bool {
	s.Errors = make([]string, 0)
	if sb.MinInstances < 1 {
		s.Errors = append(s.Errors, "Minimum number of instances must be greater than 0.")
	}
	if sb.CPUMinThreshold < 5 {
		s.Errors = append(s.Errors, "Lower CPU % threshold must be greater than or equal to 5%.")
	}
	if sb.MinInstances > sb.MaxInstances {
		s.Errors = append(s.Errors, "Minimum number of instances must be less than or equal to maximum number of instances.")
	}
	if sb.CPUMinThreshold > sb.CPUMaxThreshold {
		s.Errors = append(s.Errors, "Lower CPU % threshold must be less than Upper CPU % threshold.")
	}
	return len(s.Errors) == 0
}
