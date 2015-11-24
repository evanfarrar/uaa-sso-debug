package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/evanfarrar/uaa-sso-debug/config"

	webservices "github.com/evanfarrar/uaa-sso-debug/web/services"
	"github.com/evanfarrar/uaa-sso-debug/web/utilities"
	"github.com/ryanmoran/stack"
)

type Dashboard struct {
	scalingDecisionsRepo models.ScalingDecisionsInterface
	serviceBindingsRepo  models.ServiceBindingsInterface
	scheduledRulesRepo   models.ScheduledRulesInterface
	encryptionKey        string
}

func NewDashboard(scalingDecisionsRepo models.ScalingDecisionsInterface, serviceBindingsRepo models.ServiceBindingsInterface,
	scheduledRulesRepo models.ScheduledRulesInterface) Dashboard {
	return Dashboard{
		scalingDecisionsRepo: scalingDecisionsRepo,
		serviceBindingsRepo:  serviceBindingsRepo,
		scheduledRulesRepo:   scheduledRulesRepo,
		encryptionKey:        config.NewEnvironment().EncryptionKey,
	}
}

func (handler Dashboard) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	guid := utilities.Vars(req)["guid"]
	bindings, err := handler.serviceBindingsRepo.FindAllByServiceInstanceGuid(guid)
	if err != nil {
		panic(err)
	}

	tmpl, err := handler.getTemplate()
	if err != nil {
		panic(err)
	}

	tiles, err := handler.getAppTiles(bindings)
	if err != nil {
		panic(err)
	}

	jsonBindings, err := json.Marshal(tiles)
	if err != nil {
		panic(err)
	}
	session := webservices.NewSession([]byte(handler.encryptionKey), webservices.SessionName, req, nil)

	username, ok := session.Get("username")
	if !ok {
		panic("failed to retrieve username from session")
	}

	env := config.NewEnvironment()
	err = tmpl.Execute(w, map[string]interface{}{
		"Bindings":  string(jsonBindings),
		"User":      username,
		"LoginHost": env.LoginHost,
	})

	if err != nil {
		panic(err)
	}
}

func (handler Dashboard) getTemplate() (*template.Template, error) {
	env := config.NewEnvironment()
	tmpl := template.New("Dashboard")
	source, err := ioutil.ReadFile(filepath.Clean(env.PublicPath + "/index.html"))
	if err != nil {
		return tmpl, err
	}

	tmpl, err = tmpl.Parse(string(source))
	if err != nil {
		return tmpl, err
	}

	return tmpl, nil
}

func (handler Dashboard) getAppTiles(bindings []models.ServiceBinding) ([]models.AppTile, error) {
	tiles := []models.AppTile{}

	for _, binding := range bindings {
		appTile, err := models.NewAppTileFromBinding(binding, handler.scalingDecisionsRepo, handler.scheduledRulesRepo)
		if err != nil {
			panic(err)
		}

		tiles = append(tiles, appTile)
	}
	return tiles, nil
}
