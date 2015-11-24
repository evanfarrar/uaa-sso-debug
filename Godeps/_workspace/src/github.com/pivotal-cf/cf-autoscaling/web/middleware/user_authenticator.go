package middleware

import (
	"net/http"
	"strings"

	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/log"
	"github.com/evanfarrar/uaa-sso-debug/services"
	webservices "github.com/evanfarrar/uaa-sso-debug/web/services"
	"github.com/evanfarrar/uaa-sso-golang/uaa"
	"github.com/ryanmoran/stack"
)

type UserAuthenticator struct {
	auth services.UAAInterface
}

func NewUserAuthenticator(auth services.UAAInterface) UserAuthenticator {
	return UserAuthenticator{
		auth: auth,
	}
}

func (authenticator UserAuthenticator) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) bool {
	env := config.NewEnvironment()

	if authHeader := req.Header.Get("Authorization"); authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "bearer ")
		token := uaa.Token{Access: tokenString}
		expired, err := token.IsExpired()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return false
		}
		if expired {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		return true
	}

	session := webservices.NewSession([]byte(env.EncryptionKey), webservices.SessionName, req, w)
	accessToken, _ := session.Get("access-token")
	refreshToken, _ := session.Get("refresh-token")
	token := uaa.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}

	if token.IsPresent() {
		expired, err := token.IsExpired()
		if err != nil {
			panic(err)
		}
		if expired {
			token, err := authenticator.auth.Refresh(token.Refresh)
			if err == uaa.InvalidRefreshToken {
				log.Println("REDIRECT: Refresh token is invalid")
				return authenticator.redirectToLogin(session, w, req)
			}
			if err != nil {
				panic(err)
			}
			session.Set("access-token", token.Access)
			session.Set("refresh-token", token.Refresh)
			session.Save()
		}
	} else {
		log.Println("REDIRECT: Token is missing")
		return authenticator.redirectToLogin(session, w, req)
	}
	return true
}

func (authenticator UserAuthenticator) redirectToLogin(session webservices.Session, w http.ResponseWriter, req *http.Request) bool {
	session.Set("return-to", req.URL.Path)
	session.Save()

	http.Redirect(w, req, "/sessions/new", http.StatusFound)

	return false
}
