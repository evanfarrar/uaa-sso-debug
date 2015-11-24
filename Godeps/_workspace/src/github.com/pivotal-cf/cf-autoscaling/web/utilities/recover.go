package utilities

import (
	"errors"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/evanfarrar/uaa-sso-debug/log"

	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-golang/uaa"
	"github.com/ryanmoran/stack"
)

func Recover(w http.ResponseWriter, req *http.Request, context stack.Context, err interface{}) {
	if err != nil {
		log.PrintlnErr("[Recover]", err)
		log.PrintlnErr("[Recover]", string(debug.Stack()))

		_, isUAAFailure := err.(uaa.Failure)
		ccError := errors.New("CC ERROR")
		if err == services.CCErrors.Failure || isUAAFailure {
			err = ccError
		}

		switch err {
		case models.ErrRecordNotFound:
			errorFormatter(w, req, "", http.StatusNotFound)
		case ccError:
			errorFormatter(w, req, "We could not communicate with cloud foundry in order to get the data for this page, try again later", http.StatusInternalServerError)
		default:
			errorFormatter(w, req, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func errorFormatter(w http.ResponseWriter, req *http.Request, errorString string, status int) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		errorString = `{"errors":["` + errorString + `"]}`
	}
	w.WriteHeader(status)
	w.Write([]byte(errorString))
}
