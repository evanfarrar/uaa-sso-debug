package broker

import (
	"net/http"

	. "github.com/evanfarrar/uaa-sso-debug/controllers"
	"github.com/evanfarrar/uaa-sso-debug/services"
)

func (c Controller) Unbind(w http.ResponseWriter, req *http.Request) {
	c.withBasicAuth(c.unbind)(w, req)
}

func (c Controller) unbind(w http.ResponseWriter, req *http.Request) {
	guid := req.URL.Query().Get(":guid")
	serviceInstanceGuid := req.URL.Query().Get(":serviceInstanceGuid")

	err := c.broker.Unbind(guid, serviceInstanceGuid)
	if err != nil {
		switch err {
		case services.BrokerErrors.ServiceInstanceNotFound:
			w.WriteHeader(http.StatusNotFound)
		case services.BrokerErrors.ServiceBindingNotFound:
			w.WriteHeader(http.StatusGone)
		default:
			InternalServerError(w, req, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
