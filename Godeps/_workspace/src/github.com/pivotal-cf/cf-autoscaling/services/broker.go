package services

import (
	"errors"


)

var BrokerErrors = struct {
	DuplicateServiceInstance error
	ServiceInstanceNotFound  error
	DuplicateServiceBinding  error
	ServiceBindingNotFound   error
}{
	errors.New("Duplicate Service Instance"),
	errors.New("Service Instance Not Found"),
	errors.New("Duplicate Service Binding"),
	errors.New("Service Binding Not Found"),
}

type BrokerInterface interface {
	Provision(string, string) error
	Bind(string, string, string) error
	Unbind(string, string) error
	Deprovision(string) error
}

type Broker struct {
	serviceInstancesRepo models.ServiceInstancesInterface
	serviceBindingsRepo  models.ServiceBindingsInterface
	scheduledRulesRepo   models.ScheduledRulesInterface
}

func NewBroker(serviceInstancesRepo models.ServiceInstancesInterface, serviceBindingsRepo models.ServiceBindingsInterface, scheduledRulesRepo models.ScheduledRulesInterface) Broker {
	return Broker{
		serviceInstancesRepo: serviceInstancesRepo,
		serviceBindingsRepo:  serviceBindingsRepo,
		scheduledRulesRepo:   scheduledRulesRepo,
	}
}

func (b Broker) Provision(guid, planGuid string) error {
	_, err := b.serviceInstancesRepo.Create(models.ServiceInstance{
		Guid:     guid,
		PlanGuid: planGuid,
	})

	if err != nil {
		if err == models.ErrDuplicateRecord {
			err = BrokerErrors.DuplicateServiceInstance
		}
	}
	return err
}

func (b Broker) Bind(guid, serviceInstanceGuid, appGuid string) error {
	_, err := b.serviceInstancesRepo.Find(serviceInstanceGuid)
	if err != nil {
		if err == models.ErrRecordNotFound {
			err = BrokerErrors.ServiceInstanceNotFound
		}
		return err
	}

	serviceBindings, err := b.serviceBindingsRepo.FindAllByAppGuid(appGuid)
	if err != nil {
		return err
	}
	if len(serviceBindings) > 0 {
		return BrokerErrors.DuplicateServiceBinding
	}

	_, err = b.serviceBindingsRepo.Create(models.ServiceBinding{
		Guid:                guid,
		ServiceInstanceGuid: serviceInstanceGuid,
		AppGuid:             appGuid,
	})
	if err != nil {
		if err == models.ErrDuplicateRecord {
			err = BrokerErrors.DuplicateServiceBinding
		}
		return err
	}

	return nil
}

func (b Broker) Unbind(guid, serviceInstanceGuid string) error {
	_, err := b.serviceInstancesRepo.Find(serviceInstanceGuid)
	if err != nil {
		if err == models.ErrRecordNotFound {
			err = BrokerErrors.ServiceInstanceNotFound
		}
		return err
	}

	sb, err := b.serviceBindingsRepo.Find(guid)
	if err != nil {
		if err == models.ErrRecordNotFound {
			err = BrokerErrors.ServiceBindingNotFound
		}
		return err
	}

	_, err = b.serviceBindingsRepo.Destroy(sb)
	if err != nil {
		return err
	}

	_, err = b.scheduledRulesRepo.DeleteAllByServiceBindingGUID(guid)
	if err != nil {
		return err
	}

	return nil
}

func (b Broker) Deprovision(serviceInstanceGuid string) error {
	serviceBindings, err := b.serviceBindingsRepo.FindAllByServiceInstanceGuid(serviceInstanceGuid)
	if err != nil {
		return err
	}
	for _, sb := range serviceBindings {
		_, err = b.serviceBindingsRepo.Destroy(sb)
		if err != nil {
			return err
		}
	}

	si, err := b.serviceInstancesRepo.Find(serviceInstanceGuid)
	if err != nil {
		if err == models.ErrRecordNotFound {
			err = BrokerErrors.ServiceInstanceNotFound
		}
		return err
	}

	_, err = b.serviceInstancesRepo.Destroy(si)
	if err != nil {
		return err
	}
	return nil
}
