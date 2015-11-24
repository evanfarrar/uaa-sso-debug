package handlers

import (
	"encoding/json"
	"net/http"


	"github.com/evanfarrar/uaa-sso-debug/web/utilities"
	"github.com/ryanmoran/stack"
)

type ScheduledRulesIndex struct {
	scheduledRulesRepo models.ScheduledRulesInterface
}

func NewScheduledRulesIndex(scheduledRulesRepo models.ScheduledRulesInterface) ScheduledRulesIndex {
	return ScheduledRulesIndex{
		scheduledRulesRepo: scheduledRulesRepo,
	}
}

func (handler ScheduledRulesIndex) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	guid := utilities.Vars(req)["guid"]
	rules, err := handler.scheduledRulesRepo.FindAllByServiceBindingGuid(guid)
	if err != nil {
		panic(err)
	}

	body, err := json.Marshal(rules)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
