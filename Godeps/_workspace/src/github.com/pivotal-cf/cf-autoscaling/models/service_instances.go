package models

import (
	"database/sql"
	"strings"
	"time"
)

type ServiceInstancesInterface interface {
	Find(string) (ServiceInstance, error)
	Create(ServiceInstance) (ServiceInstance, error)
	Update(ServiceInstance) (int, error)
	Destroy(ServiceInstance) (int, error)
}

type ServiceInstances struct{}

func NewServiceInstancesRepo() ServiceInstances {
	return ServiceInstances{}
}

func (repo ServiceInstances) Find(guid string) (ServiceInstance, error) {
	si := ServiceInstance{}
	err := Database().Connection.SelectOne(&si, "SELECT * FROM `service_instances` WHERE `guid` = ?", guid)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrRecordNotFound
		}
		return si, err
	}
	return si, nil
}

func (repo ServiceInstances) Create(si ServiceInstance) (ServiceInstance, error) {
	var epoch time.Time
	if si.CreatedAt.Unix() == epoch.Unix() {
		si.CreatedAt = time.Now().Truncate(1 * time.Second).UTC()
	} else {
		si.CreatedAt = si.CreatedAt.Truncate(1 * time.Second).UTC()
	}

	err := Database().Connection.Insert(&si)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = ErrDuplicateRecord
		}
		return si, err
	}
	return si, nil
}

func (repo ServiceInstances) Update(si ServiceInstance) (int, error) {
	count, err := Database().Connection.Update(&si)
	return int(count), err
}

func (repo ServiceInstances) Destroy(si ServiceInstance) (int, error) {
	count, err := Database().Connection.Delete(&si)
	return int(count), err
}
