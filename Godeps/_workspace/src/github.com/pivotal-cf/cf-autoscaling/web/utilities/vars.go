package utilities

import (
	"net/http"

	"github.com/gorilla/mux"
)

var vars map[*http.Request]map[string]string

func init() {
	vars = make(map[*http.Request]map[string]string)
}

func SetVars(req *http.Request, v map[string]string) {
	vars[req] = v
}

func Vars(req *http.Request) map[string]string {
	if v, ok := vars[req]; ok {
		return v
	}
	return mux.Vars(req)
}
