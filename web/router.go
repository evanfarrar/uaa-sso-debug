package web

import (
	"strings"

	"github.com/gorilla/mux"
	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-debug/web/handlers"
	"github.com/ryanmoran/stack"
)

type Router struct {
	stacks map[string]stack.Stack
}

func NewRouter(auth services.UAAInterface) Router {
	handlers.NewSessionsCreate(auth)
	return Router{
	}
}

func (router Router) Routes() *mux.Router {
	r := mux.NewRouter()
	for methodPath, stack := range router.stacks {
		var name = methodPath
		parts := strings.SplitN(methodPath, " ", 2)
		r.Handle(parts[1], stack).Methods(parts[0]).Name(name)
	}
	return r
}
