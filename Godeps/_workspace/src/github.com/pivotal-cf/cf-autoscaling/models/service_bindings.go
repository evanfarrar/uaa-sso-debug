package models

import (
	"database/sql"
	"strings"
	"time"
)

type ServiceBindingsInterface interface {
	Find(string) (ServiceBinding, error)
	FindAllByServiceInstanceGuid(string) ([]ServiceBinding, error)
	FindAllEnabled() ([]ServiceBinding, error)
	FindAllByAppGuid(string) ([]ServiceBinding, error)
	Update(sb ServiceBinding) (int, error)
	Create(sb ServiceBinding) (ServiceBinding, error)
	Destroy(sb ServiceBinding) (int, error)
}

type ServiceBindings struct{}

func NewServiceBindingsRepo() ServiceBindings {
	return ServiceBindings{}
}

func (repo ServiceBindings) Find(guid string) (ServiceBinding, error) {
	sb := ServiceBinding{}
	err := Database().Connection.SelectOne(&sb, "SELECT * FROM `service_bindings` WHERE `guid` = ?", guid)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrRecordNotFound
		}
		return sb, err
	}
	return sb, nil
}

func (repo ServiceBindings) Create(sb ServiceBinding) (ServiceBinding, error) {
	sb.MinInstances = 2
	sb.MaxInstances = 5
	sb.CPUMinThreshold = 20
	sb.CPUMaxThreshold = 80
	if !sb.Enabled {
		sb.Enabled = false
	}

	var epoch time.Time
	if sb.CreatedAt.Unix() == epoch.Unix() {
		sb.CreatedAt = time.Now().Truncate(1 * time.Second).UTC()
	} else {
		sb.CreatedAt = sb.CreatedAt.Truncate(1 * time.Second).UTC()
	}

	err := Database().Connection.Insert(&sb)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = ErrDuplicateRecord
		}
		return sb, err
	}
	return sb, nil
}

func (repo ServiceBindings) Destroy(sb ServiceBinding) (int, error) {
	count, err := Database().Connection.Delete(&sb)
	return int(count), err
}

func (repo ServiceBindings) FindAllByAppGuid(appGuid string) ([]ServiceBinding, error) {
	results, err := Database().Connection.Select(ServiceBinding{}, "SELECT * FROM service_bindings WHERE app_guid = ?", appGuid)
	bindings := make([]ServiceBinding, 0)
	for _, result := range results {
		binding := result.(*ServiceBinding)
		bindings = append(bindings, *binding)
	}

	return bindings, err
}

func (repo ServiceBindings) FindAllByServiceInstanceGuid(serviceInstanceGuid string) ([]ServiceBinding, error) {
	results, err := Database().Connection.Select(ServiceBinding{}, "SELECT * FROM service_bindings WHERE service_instance_guid = ?", serviceInstanceGuid)
	bindings := make([]ServiceBinding, 0)
	for _, result := range results {
		binding := result.(*ServiceBinding)
		bindings = append(bindings, *binding)
	}

	return bindings, err
}

func (repo ServiceBindings) FindAll() ([]ServiceBinding, error) {
	results, err := Database().Connection.Select(ServiceBinding{}, "SELECT * FROM service_bindings")
	bindings := make([]ServiceBinding, 0)
	for _, result := range results {
		binding := result.(*ServiceBinding)
		bindings = append(bindings, *binding)
	}

	return bindings, err
}

func (repo ServiceBindings) FindAllEnabled() ([]ServiceBinding, error) {
	results, err := Database().Connection.Select(ServiceBinding{}, "SELECT * FROM service_bindings WHERE enabled = true")
	bindings := make([]ServiceBinding, 0)
	for _, result := range results {
		binding := result.(*ServiceBinding)
		bindings = append(bindings, *binding)
	}

	return bindings, err
}

func (repo ServiceBindings) Update(sb ServiceBinding) (int, error) {
	count, err := Database().Connection.Update(&sb)
	return int(count), err
}
