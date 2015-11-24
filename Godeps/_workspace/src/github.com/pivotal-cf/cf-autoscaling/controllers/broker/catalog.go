package broker

import (
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/evanfarrar/uaa-sso-debug/config"
	. "github.com/evanfarrar/uaa-sso-debug/controllers"
)

func (c Controller) Catalog(w http.ResponseWriter, req *http.Request) {
	c.withBasicAuth(c.catalog)(w, req)
}

func (c Controller) catalog(w http.ResponseWriter, req *http.Request) {
	env := config.NewEnvironment()
	source, err := ioutil.ReadFile(filepath.Clean(env.PublicPath + "/catalog.json"))
	if err != nil {
		InternalServerError(w, req, err)
		return
	}
	w.Write(source)
}
