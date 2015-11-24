package broker

import (
	"net/http"

	"github.com/evanfarrar/uaa-sso-debug/services"

	. "github.com/evanfarrar/uaa-sso-debug/controllers"
)

func (c Controller) Bind(w http.ResponseWriter, req *http.Request) {
	c.withBasicAuth(c.bind)(w, req)
}

func (c Controller) bind(w http.ResponseWriter, req *http.Request) {
	guid := req.URL.Query().Get(":guid")
	serviceInstanceGuid := req.URL.Query().Get(":serviceInstanceGuid")

	params, err := parseRequestBody(w, req)
	if err != nil {
		if err == ErrMalformedRequest {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var (
		app_guid string
		ok       bool
	)
	if app_guid, ok = params["app_guid"].(string); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.broker.Bind(guid, serviceInstanceGuid, app_guid)
	if err != nil {
		switch err {
		case services.BrokerErrors.ServiceInstanceNotFound:
			w.WriteHeader(http.StatusNotFound)
		case services.BrokerErrors.DuplicateServiceBinding:
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"description":"You can only bind your application to one instance of this auto-scaling service"}`))
		default:
			InternalServerError(w, req, err)
		}
		return
	}

	err = c.updateAppName(guid)
	if err != nil {
		InternalServerError(w, req, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"credentials":{}}`))
}

func (c Controller) updateAppName(guid string) error {
	binding, err := c.serviceBindingsRepo.Find(guid)
	application, err := c.CC.GetApplication(binding.AppGuid)
	if err != nil {
		return err
	}
	binding, err = c.serviceBindingsRepo.Find(binding.Guid)
	if err != nil {
		return err
	}
	binding.AppName = application.Name
	_, err = c.serviceBindingsRepo.Update(binding)
	return err
}
