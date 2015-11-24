package handlers

import (
	"encoding/json"
	"net/http"


	"github.com/evanfarrar/uaa-sso-debug/web/utilities"
	"github.com/ryanmoran/stack"
)

type ScalingDecisionsIndex struct {
	scalingDecisionsRepo models.ScalingDecisionsInterface
}

func NewScalingDecisionsIndex(scalingDecisionsRepo models.ScalingDecisionsInterface) ScalingDecisionsIndex {
	return ScalingDecisionsIndex{
		scalingDecisionsRepo: scalingDecisionsRepo,
	}
}

func (handler ScalingDecisionsIndex) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	guid := utilities.Vars(req)["guid"]
	scalingDecisions, err := handler.scalingDecisionsRepo.FindAllByServiceBindingGuid(guid)
	if err != nil {
		panic(err)
	}

	body, err := json.Marshal(scalingDecisions)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
