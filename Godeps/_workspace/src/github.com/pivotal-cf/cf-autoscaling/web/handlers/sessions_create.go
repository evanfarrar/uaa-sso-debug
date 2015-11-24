package handlers

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/services"
	webservices "github.com/evanfarrar/uaa-sso-debug/web/services"
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
	env := config.NewEnvironment()

	code := req.URL.Query().Get("code")
	token, err := handler.auth.Exchange(code)
	if err != nil {
		panic(err)
	}

	session := webservices.NewSession([]byte(env.EncryptionKey), webservices.SessionName, req, w)
	session.Set("access-token", token.Access)
	session.Set("refresh-token", token.Refresh)

	jwtToken, err := jwt.Parse(token.Access, func(_ *jwt.Token) (interface{}, error) {
		return []byte(handler.uaaPublicKey), nil
	})
	if err != nil {
		panic(err)
	}

	session.Set("username", jwtToken.Claims["user_name"].(string))
	session.Save()

	if redirectTo, ok := session.Get("return-to"); ok {
		http.Redirect(w, req, redirectTo, http.StatusFound)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not redirect"))
	}
}
