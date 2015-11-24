package handlers

import (
	"net/http"
	"strconv"


	"github.com/evanfarrar/uaa-sso-debug/web/utilities"
	"github.com/ryanmoran/stack"
)

type ScheduledRuleDelete struct {
	scheduledRulesRepo models.ScheduledRulesInterface
}

func NewScheduledRuleDelete(scheduledRulesRepo models.ScheduledRulesInterface) ScheduledRuleDelete {
	return ScheduledRuleDelete{
		scheduledRulesRepo: scheduledRulesRepo,
	}
}

func (handler ScheduledRuleDelete) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	id := utilities.Vars(req)["id"]
	scheduledRuleID, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		panic(err)
	}

	scheduledRule, err := handler.scheduledRulesRepo.Find(int(scheduledRuleID))

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

	_, err = handler.scheduledRulesRepo.Destroy(scheduledRule)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
