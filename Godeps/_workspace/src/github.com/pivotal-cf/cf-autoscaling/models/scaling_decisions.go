package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

type ScalingDecisionsInterface interface {
	FindLatestByServiceBindingGuid(string) (ScalingDecision, error)
	RecentlyNotifiedMaxInstanceReached(string) (bool, error)
	FindAllByServiceBindingGuid(string) ([]ScalingDecision, error)
	FindFirstNonNotified() (ScalingDecision, error)
	DeleteAllBefore(time.Time) (int, error)
	Create(ScalingDecision) (ScalingDecision, error)
	Update(ScalingDecision) (int, error)
}

type ScalingDecisions struct{}

func NewScalingDecisionsRepo() ScalingDecisions {
	return ScalingDecisions{}
}

func (repo ScalingDecisions) Create(scalingDecision ScalingDecision) (ScalingDecision, error) {
	var epoch time.Time
	if scalingDecision.CreatedAt.Unix() == epoch.Unix() {
		scalingDecision.CreatedAt = time.Now().Truncate(1 * time.Second).UTC()
	} else {
		scalingDecision.CreatedAt = scalingDecision.CreatedAt.Truncate(1 * time.Second).UTC()
	}
	err := Database().Connection.Insert(&scalingDecision)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = ErrDuplicateRecord
		}
		return scalingDecision, err
	}
	return scalingDecision, nil
}

func (repo ScalingDecisions) Update(scalingDecision ScalingDecision) (int, error) {
	count, err := Database().Connection.Update(&scalingDecision)
	return int(count), err
}

func (repo ScalingDecisions) DeleteAllBefore(createdAt time.Time) (int, error) {
	result, err := Database().Connection.Exec("DELETE FROM `scaling_decisions` WHERE `created_at` < ? ", createdAt)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return int(count), err
}

func (repo ScalingDecisions) Find(id int) (ScalingDecision, error) {
	scalingDecision := ScalingDecision{}
	err := Database().Connection.SelectOne(&scalingDecision, "SELECT * FROM `scaling_decisions` WHERE `id` = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrRecordNotFound
		}
		return scalingDecision, err
	}
	return scalingDecision, nil
}

func (repo ScalingDecisions) FindAllByServiceBindingGuid(serviceBindingGuid string) ([]ScalingDecision, error) {
	var scalingDecisions []ScalingDecision

	query := "SELECT * FROM `scaling_decisions` WHERE `service_binding_guid` = ? ORDER BY `created_at` DESC"

	results, err := Database().Connection.Select(ScalingDecision{}, query, serviceBindingGuid)
	for _, result := range results {
		scalingDecision := result.(*ScalingDecision)
		scalingDecisions = append(scalingDecisions, *scalingDecision)
	}

	return scalingDecisions, err
}

func (repo ScalingDecisions) FindLatestByServiceBindingGuid(serviceBindingGuid string) (ScalingDecision, error) {
	scalingDecision := ScalingDecision{}
	query := "SELECT * FROM `scaling_decisions` WHERE `service_binding_guid` = ? ORDER BY `created_at` DESC LIMIT 1"
	err := Database().Connection.SelectOne(&scalingDecision, query, serviceBindingGuid)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrRecordNotFound
		}
		return scalingDecision, err
	}
	return scalingDecision, nil
}

func (repo ScalingDecisions) FindFirstNonNotified() (ScalingDecision, error) {
	scalingDecision := ScalingDecision{}
	query := "SELECT * FROM `scaling_decisions` WHERE `notified` = 0 ORDER BY `created_at` ASC LIMIT 1"
	err := Database().Connection.SelectOne(&scalingDecision, query)

	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrRecordNotFound
		}
		return scalingDecision, err
	}

	return scalingDecision, nil
}

func (repo ScalingDecisions) RecentlyNotifiedMaxInstanceReached(serviceBindingGuid string) (bool, error) {
	query := "SELECT * FROM `scaling_decisions` WHERE `service_binding_guid` = ? ORDER BY `created_at` DESC LIMIT 2"
	results, err := Database().Connection.Select(ScalingDecision{}, query, serviceBindingGuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrRecordNotFound
		}
		return false, err
	}

	if len(results) == 2 {
		secondMostRecentDecision, ok := results[1].(*ScalingDecision)
		if !ok {
			return false, errors.New("Improperly formatted scaling decision in database")
		}

		return secondMostRecentDecision.DecisionType == FailedToScaleMaxInstanceCountReached, nil
	}
	return false, nil
}
