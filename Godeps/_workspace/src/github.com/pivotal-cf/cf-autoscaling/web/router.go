package web

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/log"

	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-debug/web/handlers"
	"github.com/evanfarrar/uaa-sso-debug/web/middleware"
	"github.com/evanfarrar/uaa-sso-debug/web/utilities"
	"github.com/ryanmoran/stack"
)

type Router struct {
	stacks map[string]stack.Stack
}

func NewRouter(auth services.UAAInterface) Router {
	env := config.NewEnvironment()

	return Router{
		stacks: map[string]stack.Stack{
			"GET /info":                                        buildUnauthenticatedStack(handlers.NewInfo()),
			"GET /dashboard/instances/{guid}":                  buildAuthenticatedStack(auth, handlers.NewDashboard(models.ScalingDecisions{}, models.ServiceBindings{}, models.ScheduledRules{})),
			"GET /api/bindings/{guid}":                         buildAuthenticatedStack(auth, handlers.NewServiceBindingsGet(models.ServiceBindings{}, models.ScalingDecisions{}, models.ScheduledRules{})),
			"POST /api/bindings/{guid}":                        buildAuthenticatedStack(auth, handlers.NewServiceBindingsUpdate(models.ServiceBindings{}, services.NewCloudControllerClient())),
			"GET /api/bindings/{guid}/decisions":               buildAuthenticatedStack(auth, handlers.NewScalingDecisionsIndex(models.ScalingDecisions{})),
			"GET /api/bindings/{guid}/scheduled_rules":         buildAuthenticatedStack(auth, handlers.NewScheduledRulesIndex(models.ScheduledRules{})),
			"POST /api/bindings/{guid}/scheduled_rules":        buildAuthenticatedStack(auth, handlers.NewScheduledRulesCreate(models.ScheduledRules{})),
			"PUT /api/bindings/{guid}/scheduled_rules/{id}":    buildAuthenticatedStack(auth, handlers.NewScheduledRulesUpdate(models.ScheduledRules{})),
			"DELETE /api/bindings/{guid}/scheduled_rules/{id}": buildAuthenticatedStack(auth, handlers.NewScheduledRuleDelete(models.ScheduledRules{})),
			"GET /sessions/new":                                buildUnauthenticatedStack(handlers.NewSessionsNew(auth)),
			"GET /sessions/create":                             buildUnauthenticatedStack(handlers.NewSessionsCreate(auth)),
			"GET /public/{path:.*}":                            buildUnauthenticatedStack(stack.CompatibleHandler(http.StripPrefix("/public/", http.FileServer(http.Dir(env.PublicPath))))),
		},
	}
}

func buildAuthenticatedStack(auth services.UAAInterface, handler stack.Handler) stack.Stack {
	logging := stack.NewLogging(log.Logger)
	authenticator := middleware.NewUserAuthenticator(auth)
	permissions := middleware.NewPermissions(models.ServiceBindings{}, services.NewCloudControllerClient())
	return DefaultStack(handler).Use(logging, authenticator, permissions)
}

func buildUnauthenticatedStack(handler stack.Handler) stack.Stack {
	logging := stack.NewLogging(log.Logger)
	return DefaultStack(handler).Use(logging)
}

func DefaultStack(handler stack.Handler) stack.Stack {
	s := stack.NewStack(handler)
	callback := stack.RecoverCallback(utilities.Recover)
	s.RecoverCallback = &callback

	return s
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
