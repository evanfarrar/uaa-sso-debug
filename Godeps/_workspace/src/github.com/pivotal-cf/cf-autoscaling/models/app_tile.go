package models

import (
	"time"
)

type AppTile struct {
	Guid                    string    `json:"guid"`
	AppName                 string    `json:"app_name"`
	MinInstances            int       `json:"min_instances"`
	MaxInstances            int       `json:"max_instances"`
	CPUMinThreshold         int       `json:"cpu_min_threshold"`
	CPUMaxThreshold         int       `json:"cpu_max_threshold"`
	Enabled                 bool      `json:"enabled"`
	ScalingEventDescription string    `json:"scaling_event_description"`
	ScalingEventTime        time.Time `json:"scaling_event_time"`
	ScheduledRulesCount     int       `json:"scheduled_rules_count"`
	NextScheduledEventTime  string    `json:"next_scheduled_event_time"`
}

func NewAppTileFromBinding(binding ServiceBinding,
	decisionsRepo ScalingDecisionsInterface,
	rulesRepo ScheduledRulesInterface) (AppTile, error) {

	rulesCount, err := rulesRepo.CountByServiceBindingGuid(binding.Guid)
	if err != nil {
		return AppTile{}, err
	}

	appTile := AppTile{
		Guid:                    binding.Guid,
		AppName:                 binding.AppName,
		MinInstances:            binding.MinInstances,
		MaxInstances:            binding.MaxInstances,
		CPUMinThreshold:         binding.CPUMinThreshold,
		CPUMaxThreshold:         binding.CPUMaxThreshold,
		Enabled:                 binding.Enabled,
		ScalingEventDescription: "NOT FOUND",
		ScalingEventTime:        time.Now(),
		ScheduledRulesCount:     rulesCount,
		NextScheduledEventTime:  "No Upcoming Events",
	}

	decision, err := decisionsRepo.FindLatestByServiceBindingGuid(binding.Guid)
	if err != nil {
		if err != ErrRecordNotFound {
			return appTile, err
		}
	} else {
		appTile.ScalingEventDescription = decision.Description
		appTile.ScalingEventTime = decision.CreatedAt
	}

	event, err := rulesRepo.NextScheduledEventByServiceBindingGuid(binding.Guid)
	if err != nil {
		if err != ErrNoScheduledEventsFound {
			return appTile, err
		}
	} else {
		appTile.NextScheduledEventTime = event.ExecutesAt.Format(time.RFC3339)
	}

	return appTile, nil
}
