package validators

import (
	"time"


)

type ScheduledRuleValidator struct {
	Errors []string
	Repo   models.ScheduledRulesInterface
}

func NewScheduledRuleValidator(repo models.ScheduledRulesInterface) ScheduledRuleValidator {
	return ScheduledRuleValidator{
		Repo: repo,
	}
}

func (s *ScheduledRuleValidator) Validate(sr models.ScheduledRule) bool {
	s.Errors = make([]string, 0)

	if sr.IsNew() && sr.ExecutesAt.Before(time.Now()) {
		s.Errors = append(s.Errors, "Rules cannot be set in the past.")
	}

	if sr.MinInstances < 1 {
		s.Errors = append(s.Errors, "Minimum number of instances must be greater than 0.")
	}

	if sr.MinInstances > sr.MaxInstances {
		s.Errors = append(s.Errors, "Minimum number of instances must be less than or equal to maximum number of instances.")
	}

	if sr.IsRecurring() {
		recurringRules, err := s.Repo.FindAllRecurringByServiceBindingGuid(sr.ServiceBindingGuid)
		if err != nil {
			panic(err)
		}

		for _, rule := range recurringRules {
			if sr.Overlaps(rule) && sr.ID != rule.ID {
				s.Errors = append(s.Errors, "Rules cannot conflict.")
				break
			}
		}
	} else {
		conflictingRules, err := s.Repo.FindAllNonRecurringByServiceBindingGuidAndExecutesAt(sr.ServiceBindingGuid, sr.ExecutesAt)
		if err != nil {
			panic(err)
		}
		for _, rule := range conflictingRules {
			if sr.ID != rule.ID {
				s.Errors = append(s.Errors, "Rules cannot conflict.")
				break
			}
		}
	}

	return len(s.Errors) == 0
}
