package models

import (
	"database/sql"
	"strings"
	"time"
)

type ReadingsInterface interface {
	Create(Reading) (Reading, error)
	DeleteAllBefore(time.Time) (int, error)
	Find(int) (Reading, error)
}

type Readings struct{}

func NewReadingsRepo() Readings {
	return Readings{}
}

func (repo Readings) Create(reading Reading) (Reading, error) {
	reading.CreatedAt = reading.CreatedAt.Truncate(1 * time.Second).UTC()
	err := Database().Connection.Insert(&reading)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = ErrDuplicateRecord
		}
		return reading, err
	}
	return reading, nil
}

func (repo Readings) Find(id int) (Reading, error) {
	reading := Reading{}
	err := Database().Connection.SelectOne(&reading, "SELECT * FROM `readings` WHERE `id` = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrRecordNotFound
		}
		return reading, err
	}
	return reading, nil
}

func (repo Readings) FindAllByServiceBindingGuid(serviceBindingGuid string) ([]Reading, error) {
	results, err := Database().Connection.Select(Reading{}, "SELECT * FROM `readings` WHERE `service_binding_guid` = ?", serviceBindingGuid)
	readings := make([]Reading, 0)
	for _, result := range results {
		reading := result.(*Reading)
		readings = append(readings, *reading)
	}

	return readings, err
}

func (repo Readings) DeleteAllBefore(createdAt time.Time) (int, error) {
	result, err := Database().Connection.Exec("DELETE FROM `readings` WHERE `created_at` < ? ", createdAt)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return int(count), err
}

func (repo Readings) FindMostRecentReadingByServiceBindingGuid(serviceBindingGuid string) (Reading, error) {
	reading := Reading{}
	query := "SELECT * FROM `readings` WHERE `service_binding_guid` = ? ORDER BY `created_at` DESC LIMIT 1"
	err := Database().Connection.SelectOne(&reading, query, serviceBindingGuid)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrRecordNotFound
		}
		return reading, err
	}
	return reading, nil
}

func (repo Readings) FindAllByServiceBindingGuidSinceTime(serviceBindingGuid string, since time.Time) ([]Reading, error) {
	query := "SELECT * FROM `readings` WHERE `service_binding_guid` = ? AND `created_at` > ? ORDER BY `created_at` DESC"
	results, err := Database().Connection.Select(Reading{}, query, serviceBindingGuid, since)
	readings := make([]Reading, 0)
	for _, result := range results {
		reading := result.(*Reading)
		readings = append(readings, *reading)
	}

	return readings, err
}

func (repo Readings) Destroy(reading Reading) (int, error) {
	count, err := Database().Connection.Delete(&reading)
	return int(count), err
}
