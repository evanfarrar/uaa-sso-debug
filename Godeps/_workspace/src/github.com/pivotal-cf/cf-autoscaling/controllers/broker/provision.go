package broker

import (
	"encoding/json"
	"net/http"

	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/services"
)

func (c Controller) Provision(w http.ResponseWriter, req *http.Request) {
	c.withBasicAuth(c.provision)(w, req)
}

func (c Controller) provision(w http.ResponseWriter, req *http.Request) {
	guid := req.URL.Query().Get(":serviceInstanceGuid")

	params, err := parseRequestBody(w, req)
	if err != nil {
		if err == ErrMalformedRequest {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var (
		plan_id string
		ok      bool
	)
	if plan_id, ok = params["plan_id"].(string); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.broker.Provision(guid, plan_id)
	if err != nil {
		if err == services.BrokerErrors.DuplicateServiceInstance {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("{}"))
			return
		}
	}

	w.WriteHeader(http.StatusCreated)

	env := config.NewEnvironment()
	result := map[string]string{
		"dashboard_url": env.Scheme + "://" + c.host + "/dashboard/instances/" + guid,
	}

	json.NewEncoder(w).Encode(result)
}
