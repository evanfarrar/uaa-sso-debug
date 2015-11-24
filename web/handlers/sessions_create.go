package handlers

import (
	"net/http"

	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/ryanmoran/stack"
)

type SessionsCreate struct {
	auth         services.UAAInterface
	uaaPublicKey string
}

func NewSessionsCreate(auth services.UAAInterface) SessionsCreate {
	key, err := auth.GetTokenKey()
	if err != nil {
		panic(err)
	}

	return SessionsCreate{auth, key}
}

func (handler SessionsCreate) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
}
