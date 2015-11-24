package application

import (
	"time"

	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/log"
	"github.com/evanfarrar/uaa-sso-debug/web"
	"github.com/evanfarrar/uaa-sso-golang/uaa"
	"github.com/ryanmoran/viron"
)

type Application struct{}

func NewApplication() Application {
	return Application{}
}

func (app Application) PrintEnvironment() {
	env := config.NewEnvironment()
	viron.Print(env, log.Logger)
}


func (app Application) UAAClient(env config.Environment) uaa.UAA {
	auth := uaa.NewUAA(env.LoginHost, env.UAAHost, env.UAAClientID, env.UAAClientSecret, "")
	auth.RedirectURL = env.Domain + "/sessions/create"
	auth.Scope = "openid,cloud_controller.permissions,cloud_controller.read,cloud_controller.write"
	auth.AccessType = "offline"
	auth.VerifySSL = env.VerifySSL

	return auth
}

func (app Application) StartServer() {
	server := web.NewServer()
	server.Start()
}

func (app Application) Crash() {
	err := recover()
	if err != nil {
		time.Sleep(10 * time.Second)
		panic(err)
	}
}
