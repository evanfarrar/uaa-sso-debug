package handlers

import (
	"encoding/json"
	"net/http"


	"github.com/evanfarrar/uaa-sso-debug/web/utilities"
	"github.com/ryanmoran/stack"
)

type ServiceBindingsGet struct {
	bindingsRepo  models.ServiceBindingsInterface
	decisionsRepo models.ScalingDecisionsInterface
	rulesRepo     models.ScheduledRulesInterface
}

func NewServiceBindingsGet(bindingsRepo models.ServiceBindingsInterface,
	decisionsRepo models.ScalingDecisionsInterface,
	rulesRepo models.ScheduledRulesInterface) ServiceBindingsGet {

	return ServiceBindingsGet{
		bindingsRepo:  bindingsRepo,
		decisionsRepo: decisionsRepo,
		rulesRepo:     rulesRepo,
	}
}

func (handler ServiceBindingsGet) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	serviceBindingGuid := utilities.Vars(req)["guid"]
	binding, err := handler.bindingsRepo.Find(serviceBindingGuid)

	if err != nil {
		if err == models.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("{}"))
			return
		}
		if err != nil {
			panic(err)
		}
	}

	appTile, err := models.NewAppTileFromBinding(binding, handler.decisionsRepo, handler.rulesRepo)
	if err != nil {
		panic(err)
	}

	body, err := json.Marshal(appTile)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
