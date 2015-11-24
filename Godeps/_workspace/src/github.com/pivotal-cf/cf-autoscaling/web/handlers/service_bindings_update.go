package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"



	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-debug/web/utilities"
	"github.com/ryanmoran/stack"
)

type ServiceBindingsUpdate struct {
	ServiceBindings models.ServiceBindingsInterface
	CloudController services.CloudControllerInterface
}

func NewServiceBindingsUpdate(bindings models.ServiceBindingsInterface, cc services.CloudControllerInterface) ServiceBindingsUpdate {
	return ServiceBindingsUpdate{
		ServiceBindings: bindings,
		CloudController: cc,
	}
}

func (handler ServiceBindingsUpdate) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	guid := utilities.Vars(req)["guid"]
	theBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	sb, err := handler.ServiceBindings.Find(guid)
	if err != nil {
		panic(err)
	}

	bindingWasEnabled := sb.Enabled

	err = json.Unmarshal(theBody, &sb)
	if err != nil {
		panic(err)
	}

	if sb.Enabled && !bindingWasEnabled {
		application, err := handler.CloudController.Stats(sb.AppGuid)
		if err != nil {
			panic(err)
		}
		sb.ExpectedInstanceCount = application.ExpectedInstanceCount
	}

	validator := validators.NewServiceBindingValidator()
	if !validator.Validate(sb) {
		w.WriteHeader(422)
		response, err := json.Marshal(map[string][]string{
			"errors": validator.Errors,
		})
		if err != nil {
			panic(err)
		}
		w.Write(response)
		return
	}

	_, err = handler.ServiceBindings.Update(sb)
	if err != nil {
		panic(err)
	}

	w.Write([]byte("{}"))
}
