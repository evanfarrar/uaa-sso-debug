package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"



	"github.com/ryanmoran/stack"
)

type ScheduledRulesCreate struct {
	rulesRepo models.ScheduledRulesInterface
}

func NewScheduledRulesCreate(rulesRepo models.ScheduledRulesInterface) ScheduledRulesCreate {
	return ScheduledRulesCreate{
		rulesRepo: rulesRepo,
	}
}

func (handler ScheduledRulesCreate) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	rule := models.ScheduledRule{}
	rule.Enabled = true

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

	rule, err = handler.rulesRepo.Create(rule)
	if err != nil {
		panic(err)
	}

	ruleJson, err := json.Marshal(rule)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(ruleJson)
}

func RenderError(w http.ResponseWriter, code int, errors []string) {
	response, err := json.Marshal(map[string][]string{
		"errors": errors,
	})
	if err != nil {
		panic(err)
	}
	w.WriteHeader(code)
	w.Write(response)
}
