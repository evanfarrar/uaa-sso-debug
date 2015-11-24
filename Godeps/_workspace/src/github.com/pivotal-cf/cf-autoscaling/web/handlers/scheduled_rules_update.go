package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"



	"github.com/evanfarrar/uaa-sso-debug/web/utilities"
	"github.com/ryanmoran/stack"
)

type ScheduledRulesUpdate struct {
	rulesRepo models.ScheduledRulesInterface
}

func NewScheduledRulesUpdate(rulesRepo models.ScheduledRulesInterface) ScheduledRulesUpdate {
	return ScheduledRulesUpdate{
		rulesRepo: rulesRepo,
	}
}

func (handler ScheduledRulesUpdate) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	ruleID, err := strconv.ParseInt(utilities.Vars(req)["id"], 10, 0)
	if err != nil {
		panic(err)
	}

	rule, err := handler.rulesRepo.Find(int(ruleID))
	if err == models.ErrRecordNotFound {
		RenderError(w, 404, []string{"The id referenced is not in the database"})
		return
	} else {
		if err != nil {
			panic(err)
		}
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &rule)
	if err != nil {
		if _, ok := err.(*time.ParseError); ok {
			RenderError(w, 422, []string{"Date and time could not be parsed."})
			return
		}
	}
	if err != nil {
		panic(err)
	}

	validator := validators.NewScheduledRuleValidator(handler.rulesRepo)
	if !validator.Validate(rule) {
		RenderError(w, 422, validator.Errors)
		return
	}

	_, err = handler.rulesRepo.Update(rule)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
