package services

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/log"
	"github.com/evanfarrar/uaa-sso-golang/uaa"
)

const (
	InstanceRunning = "RUNNING"
	AppRunning      = "RUNNING"
	AppUnstable     = "UNSTABLE"
	AppStopped      = "STOPPED"
)

var CCErrors = struct {
	Failure              error
	AppMissing           error
	AppQuotaLimitReached error
}{
	errors.New("CloudController Failure"),
	errors.New("CloudController AppMissing"),
	errors.New("CloudController AppQuotaLimitReached"),
}

type CloudControllerInterface interface {
	Stats(string) (ApplicationStats, error)
	CanManageInstance(uaa.Token, string) (bool, error)
	AppSummary(string) (ApplicationSummary, error)
}

type CloudController struct {
	CCHost string
	auth   uaa.UAA
}

type AppInstanceResponse struct {
	State string
	Stats struct {
		Usage struct {
			CPU float64
		}
	}
}

type AppResponse struct {
	Entity struct {
		Name      string
		Instances int
	}
}

type ApplicationStats struct {
	Name                  string
	ExpectedInstanceCount int
	RunningInstanceCount  int
	CPUUtilization        int
	State                 string
}

type ApplicationSummary struct {
	SpaceGUID string `json:"space_guid"`
}

func NewCloudControllerClient() CloudController {
	env := config.NewEnvironment()
	auth := uaa.NewUAA(env.LoginHost, env.UAAHost, env.UAAClientID, env.UAAClientSecret, "")
	auth.VerifySSL = env.VerifySSL
	cc := CloudController{
		CCHost: env.CCHost,
		auth:   auth,
	}
	return cc
}

func (c CloudController) CanManageInstance(token uaa.Token, serviceInstanceGuid string) (bool, error) {
	code, body, err := c.get("/v2/service_instances/"+serviceInstanceGuid+"/permissions", token)
	if err != nil {
		return false, err
	}

	if code == 401 {
		token, err = c.auth.Refresh(token.Refresh)
		if err != nil {
			return false, err
		}
		code, body, err = c.get("/v2/service_instances/"+serviceInstanceGuid+"/permissions", token)
	}

	if code > 399 {
		log.Printf("CODE, BODY -----> %+v %+v\n", code, string(body))
		return false, CCErrors.Failure
	}

	canManage := map[string]bool{}
	json.Unmarshal(body, &canManage)
	return canManage["manage"], nil
}

func (c CloudController) AppSummary(appGUID string) (ApplicationSummary, error) {
	applicationSummary := ApplicationSummary{}

	token, err := c.auth.GetClientToken()
	if err != nil {
		return applicationSummary, err
	}

	code, body, err := c.get("/v2/apps/"+appGUID+"/summary", token)
	if err != nil {
		return applicationSummary, err
	}

	if code > 399 {
		log.Printf("CODE, BODY -----> %+v %+v\n", code, string(body))
		if strings.Contains(string(body), "CF-AppNotFound") {
			return applicationSummary, CCErrors.AppMissing
		}

		return applicationSummary, CCErrors.Failure
	}

	err = json.Unmarshal(body, &applicationSummary)
	if err != nil {
		panic(err)
	}

	return applicationSummary, nil
}

func (c CloudController) Stats(appGuid string) (ApplicationStats, error) {
	stats, err := c.GetApplication(appGuid)
	if err != nil {
		return stats, err
	}

	token, err := c.auth.GetClientToken()
	if err != nil {
		return stats, err
	}
	code, body, err := c.get("/v2/apps/"+appGuid+"/stats", token)
	if err != nil {
		return stats, err
	}
	if code > 399 {
		log.Printf("CODE, BODY -----> %+v %+v\n", code, string(body))
		return stats, CCErrors.Failure
	}

	appInstances := map[string]AppInstanceResponse{}
	json.Unmarshal(body, &appInstances)

	cpuTotal := 0
	runningInstanceStats := []AppInstanceResponse{}
	for _, appInstance := range appInstances {
		cpuTotal += int(appInstance.Stats.Usage.CPU * 100)
		if appInstance.State == InstanceRunning {
			runningInstanceStats = append(runningInstanceStats, appInstance)
		} else {
			stats.State = AppUnstable
		}
	}
	if stats.State != AppUnstable {
		stats.State = AppRunning
	}
	if len(appInstances) == 0 {
		stats.State = AppStopped
	}

	stats.ExpectedInstanceCount = len(appInstances)
	if len(runningInstanceStats) > 0 {
		stats.CPUUtilization = cpuTotal / len(runningInstanceStats)
		stats.RunningInstanceCount = len(runningInstanceStats)
	}

	return stats, nil
}

func (c CloudController) GetApplication(appGuid string) (ApplicationStats, error) {
	application := ApplicationStats{}

	token, err := c.auth.GetClientToken()
	if err != nil {
		return application, err
	}
	code, body, err := c.get("/v2/apps/"+appGuid, token)
	if err != nil {
		return application, err
	}

	if code > 399 {
		log.Printf("CODE, BODY -----> %+v %+v\n", code, string(body))
		return application, CCErrors.Failure
	}

	app := AppResponse{}
	json.Unmarshal(body, &app)

	application.Name = app.Entity.Name

	return application, nil
}

func (c CloudController) Scale(appGuid string, newInstanceCount int) (int, error) {
	token, err := c.auth.GetClientToken()
	if err != nil {
		return newInstanceCount, err
	}
	requestBody, err := json.Marshal(map[string]int{"instances": newInstanceCount})
	code, body, err := c.put("/v2/apps/"+appGuid, token, string(requestBody))
	if err != nil {
		return newInstanceCount, err
	}

	if code > 399 {
		log.Printf("CODE, BODY -----> %+v %+v\n", code, string(body))
		if strings.Contains(string(body), "CF-AppMemoryQuotaExceeded") {
			return newInstanceCount, CCErrors.AppQuotaLimitReached
		}

		return newInstanceCount, CCErrors.Failure
	}

	app := AppResponse{}
	json.Unmarshal(body, &app)
	return app.Entity.Instances, nil
}

func (c CloudController) get(path string, token uaa.Token) (int, []byte, error) {
	env := config.NewEnvironment()
	client := Client{
		Host:        env.CCHost,
		ContentType: "application/json",
	}
	return client.MakeRequest("GET", path, token.Access, nil)
}

func (c CloudController) put(path string, token uaa.Token, requestBody string) (int, []byte, error) {
	env := config.NewEnvironment()
	client := Client{
		Host:        env.CCHost,
		ContentType: "application/json",
	}
	return client.MakeRequest("PUT", path, token.Access, strings.NewReader(requestBody))
}
