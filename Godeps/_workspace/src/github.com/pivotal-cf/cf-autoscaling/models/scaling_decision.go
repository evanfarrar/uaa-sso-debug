package models

import (
	"time"
)

const (
	UnknownType = iota
	FailedToScaleMaxInstanceCountReached
	ScaleCPUAboveThreshold
	NoScaleAppWithinThresholds
	ScaleDownBelowCPUUtilizationThreshold
	NoScaleAtMinimumInstanceCount
	ScalingDisabled
	ManualScaleDetected
	ScaleUpAppBelowMinAppInstanceCount
	ScaleDownAppAboveMaxInstanceCount
	FailedToScaleQuotaDisallows
)

type ScalingDecision struct {
	ID                 int       `db:"id"                    json:"id"`
	ReadingID          int       `db:"reading_id"            json:"reading_id"`
	ServiceBindingGuid string    `db:"service_binding_guid"  json:"service_binding_guid"`
	ScalingFactor      int       `db:"scaling_factor"        json:"scaling_factor"`
	Description        string    `db:"description"           json:"description"`
	CreatedAt          time.Time `db:"created_at"            json:"created_at"`
	Notified           bool      `json:"-"`
	DecisionType       int       `db:"decision_type"         json:"-"`
}

func (model ScalingDecision) IsScheduled() bool {
	return model.ReadingID == 0
}

func (model ScalingDecision) IsIgnored() bool {
	return model.DecisionType == NoScaleAtMinimumInstanceCount ||
		model.DecisionType == NoScaleAppWithinThresholds ||
		model.DecisionType == ScaleDownAppAboveMaxInstanceCount ||
		model.DecisionType == ScaleUpAppBelowMinAppInstanceCount ||
		model.DecisionType == ScalingDisabled
}
