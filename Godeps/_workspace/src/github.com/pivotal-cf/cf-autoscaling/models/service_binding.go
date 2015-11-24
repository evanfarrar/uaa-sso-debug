package models

import "time"

type ServiceBinding struct {
	Guid                  string    `db:"guid"                    json:"-"`
	ServiceInstanceGuid   string    `db:"service_instance_guid"   json:"-"`
	AppGuid               string    `db:"app_guid"                json:"app_guid"`
	AppName               string    `db:"app_name"                json:"-"`
	ExpectedInstanceCount int       `db:"expected_instance_count" json:"-"`
	MinInstances          int       `db:"min_instances"           json:"min_instances"`
	MaxInstances          int       `db:"max_instances"           json:"max_instances"`
	CPUMinThreshold       int       `db:"cpu_min_threshold"       json:"cpu_min_threshold"`
	CPUMaxThreshold       int       `db:"cpu_max_threshold"       json:"cpu_max_threshold"`
	Enabled               bool      `db:"enabled"                 json:"enabled"`
	CreatedAt             time.Time `db:"created_at"              json:"created_at"`
}
