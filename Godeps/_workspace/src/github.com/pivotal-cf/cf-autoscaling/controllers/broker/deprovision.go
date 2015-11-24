package broker

import (
	"net/http"

	. "github.com/evanfarrar/uaa-sso-debug/controllers"
	"github.com/evanfarrar/uaa-sso-debug/services"
)

func (c Controller) Deprovision(w http.ResponseWriter, req *http.Request) {
	c.withBasicAuth(c.deprovision)(w, req)
}

func (c Controller) deprovision(w http.ResponseWriter, req *http.Request) {
	serviceInstanceGuid := req.URL.Query().Get(":serviceInstanceGuid")

	err := c.broker.Deprovision(serviceInstanceGuid)
	if err != nil {
		if err == services.BrokerErrors.ServiceInstanceNotFound {
			w.WriteHeader(http.StatusGone)
			return
		} else {
			InternalServerError(w, req, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
