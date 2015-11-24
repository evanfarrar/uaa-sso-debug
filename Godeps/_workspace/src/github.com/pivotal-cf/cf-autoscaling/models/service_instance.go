package models

import "time"

type ServiceInstance struct {
	Guid      string    `db:"guid"`
	PlanGuid  string    `db:"plan_guid"       json:"plan_guid"`
	CreatedAt time.Time `db:"created_at"    json:"created_at"`
}
