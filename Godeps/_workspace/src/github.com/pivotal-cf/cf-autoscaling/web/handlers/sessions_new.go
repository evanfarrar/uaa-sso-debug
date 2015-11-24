package handlers

import (
	"net/http"

	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/ryanmoran/stack"
)

type SessionsNew struct {
	auth services.UAAInterface
}

func NewSessionsNew(auth services.UAAInterface) SessionsNew {
	return SessionsNew{
		auth: auth,
	}
}

func (handler SessionsNew) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	http.Redirect(w, req, handler.auth.LoginURL(), http.StatusFound)
}
