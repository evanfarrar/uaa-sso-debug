package broker

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/pat"
	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/services"

	. "github.com/evanfarrar/uaa-sso-debug/controllers"
)

var ErrMalformedRequest = errors.New("Malformed request body")

type Controller struct {
	host                string
	username            string
	password            string
	realm               string
	broker              services.BrokerInterface
	CC                  services.CloudController
	serviceBindingsRepo models.ServiceBindingsInterface
}

func New(broker services.BrokerInterface, serviceBindingsRepo models.ServiceBindingsInterface) Controller {
	env := config.NewEnvironment()
	return Controller{
		host:                env.Host,
		username:            env.BasicAuthUsername,
		password:            env.BasicAuthPassword,
		realm:               "autoscale",
		broker:              broker,
		CC:                  services.NewCloudControllerClient(),
		serviceBindingsRepo: serviceBindingsRepo,
	}
}

func (c Controller) Handler() http.Handler {
	handler := pat.New()
	handler.Get("/v2/catalog", c.Catalog)
	handler.Put("/v2/service_instances/{serviceInstanceGuid}/service_bindings/{guid}", c.Bind)
	handler.Put("/v2/service_instances/{serviceInstanceGuid}", c.Provision)
	handler.Delete("/v2/service_instances/{serviceInstanceGuid}/service_bindings/{guid}", c.Unbind)
	handler.Delete("/v2/service_instances/{serviceInstanceGuid}", c.Deprovision)
	return handler
}

func (c Controller) withBasicAuth(callback http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		basicAuthHeader := req.Header.Get("Authorization")
		if len(basicAuthHeader) > 0 {
			str := strings.Split(basicAuthHeader, " ")[1]
			data, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				panic(err)
			}

			credentials := strings.Split(string(data), ":")
			if c.username == credentials[0] && c.password == credentials[1] {
				callback(w, req)
			} else {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
		} else {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", c.realm))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func parseRequestBody(w http.ResponseWriter, req *http.Request) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		InternalServerError(w, req, err)
	}

	err = json.Unmarshal(body, &params)
	if err != nil {
		return params, ErrMalformedRequest
	}
	return params, nil
}
