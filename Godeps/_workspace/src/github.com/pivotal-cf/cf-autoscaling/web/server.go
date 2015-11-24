package web

import (
	"net/http"

	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/controllers/broker"
	"github.com/evanfarrar/uaa-sso-debug/log"

	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-golang/uaa"
)

type Server struct {
	host string
	port string
}

func NewServer() Server {
	env := config.NewEnvironment()
	return Server{
		host: env.Host,
		port: env.Port,
	}
}

func (s Server) Start() {
	mux := http.NewServeMux()
	mux.Handle("/v2/", broker.New(services.NewBroker(models.ServiceInstances{}, models.ServiceBindings{}, models.ScheduledRules{}), models.ServiceBindings{}).Handler())
	mux.Handle("/", NewRouter(makeAuth()).Routes())
	s.Run(mux)
}

func (s Server) Run(mux http.Handler) {
	log.Printf("Listening on %s:%s\n", s.host, s.port)
	err := http.ListenAndServe(":"+s.port, mux)
	if err != nil {
		panic(err)
	}
}

func makeAuth() *uaa.UAA {
	env := config.NewEnvironment()
	auth := uaa.NewUAA(env.LoginHost, env.UAAHost, env.UAAClientID, env.UAAClientSecret, "")
	auth.RedirectURL = env.Domain + "/sessions/create"
	auth.Scope = "openid,cloud_controller.permissions,cloud_controller.read,cloud_controller.write"
	auth.AccessType = "offline"
	auth.VerifySSL = env.VerifySSL
	return &auth
}
