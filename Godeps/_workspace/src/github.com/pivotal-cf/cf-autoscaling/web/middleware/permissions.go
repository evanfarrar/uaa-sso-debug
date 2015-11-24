package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/evanfarrar/uaa-sso-debug/config"

	"github.com/evanfarrar/uaa-sso-debug/services"
	webservices "github.com/evanfarrar/uaa-sso-debug/web/services"
	"github.com/evanfarrar/uaa-sso-golang/uaa"
	"github.com/ryanmoran/stack"
)

type Permissions struct {
	cc           services.CloudControllerInterface
	bindingsRepo models.ServiceBindingsInterface
}

func NewPermissions(bindingsRepo models.ServiceBindingsInterface, cc services.CloudControllerInterface) Permissions {
	return Permissions{
		cc:           cc,
		bindingsRepo: bindingsRepo,
	}
}

func (p Permissions) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) bool {
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
	siGuid := p.ParseGuid(req.URL.Path)
	success, err := p.cc.CanManageInstance(token, siGuid)
	if err != nil {
		panic(err)
	}

	if !success {
		w.WriteHeader(http.StatusUnauthorized)
	}

	return success
}

func (p Permissions) ParseGuid(path string) string {
	dashboardRegexp := regexp.MustCompile(`/dashboard/instances/([A-Za-z0-9\-]*)`)
	apiBindingsRegexp := regexp.MustCompile(`/api/bindings/([A-Za-z0-9\-]*)`)
	switch {
	case dashboardRegexp.MatchString(path):
		matches := dashboardRegexp.FindStringSubmatch(path)
		return matches[1]
	case apiBindingsRegexp.MatchString(path):
		matches := apiBindingsRegexp.FindStringSubmatch(path)
		binding, err := p.bindingsRepo.Find(matches[1])
		if err != nil {
			panic(err)
		}
		return binding.ServiceInstanceGuid
	}
	return ""
}
