package controllers

import (
	"net/http"
	"runtime/debug"

	"github.com/evanfarrar/uaa-sso-debug/log"
)

func InternalServerError(w http.ResponseWriter, req *http.Request, err error) {
	log.Print(err)
	log.Print(string(debug.Stack()))
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}
