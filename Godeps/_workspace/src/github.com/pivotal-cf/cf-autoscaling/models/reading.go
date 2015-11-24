package models

import (
	"time"

	"github.com/xeonx/timeago"
)

type Reading struct {
	ID                    int       `db:"id"`
	ServiceBindingGuid    string    `db:"service_binding_guid"`
	AppName               string    `db:"-"`
	CPUUtilization        int       `db:"cpu_utilization"`
	ExpectedInstanceCount int       `db:"expected_instance_count"`
	RunningInstanceCount  int       `db:"running_instance_count"`
	State                 string    `db:"state"`
	CreatedAt             time.Time `db:"created_at"`
}

func (reading Reading) TimeAgo() string {
	return timeago.English.Format(reading.CreatedAt)
}

func (reading Reading) DoesNotMatchExpectedInstanceCount(binding ServiceBinding) bool {
	return reading.ExpectedInstanceCount != binding.ExpectedInstanceCount
}
