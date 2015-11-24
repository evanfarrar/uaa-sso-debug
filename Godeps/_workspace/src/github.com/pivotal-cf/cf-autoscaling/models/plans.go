package models

import "time"

const (
	TestPlanGUID   = "test-plan"
	GoldPlanGUID   = "c23833b3-fe27-4e30-aa72-ecffc7257b70"
	BronzePlanGUID = "e4518390-ab55-412c-b22c-55c31f25db90"
)

var PlansStore = make(map[string]Plan)

type PlansInterface interface {
	Create(Plan) (Plan, error)
	Find(string) (Plan, error)
}

type Plans struct{}

func NewPlansRepo() Plans {
	return Plans{}
}

func init() {
	repo := Plans{}
	repo.Create(Plan{
		Guid:            BronzePlanGUID,
		PollingInterval: 5 * time.Minute,
	})
	repo.Create(Plan{
		Guid:            GoldPlanGUID,
		PollingInterval: 30 * time.Second,
	})

	// This plan is not used in production
	repo.Create(Plan{
		Guid:            TestPlanGUID,
		PollingInterval: 1 * time.Second,
	})
}

func (repo Plans) Create(plan Plan) (Plan, error) {
	PlansStore[plan.Guid] = plan
	return PlansStore[plan.Guid], nil
}

func (repo Plans) Find(guid string) (Plan, error) {
	plan, ok := PlansStore[guid]
	if !ok {
		return plan, ErrRecordNotFound
	}
	return plan, nil
}
