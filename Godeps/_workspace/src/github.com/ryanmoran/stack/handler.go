package stack

import "net/http"

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, Context)
}

type compatibleHandler struct {
	handler http.Handler
}

func CompatibleHandler(handler http.Handler) compatibleHandler {
	return compatibleHandler{
		handler: handler,
	}
}

func (handler compatibleHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, context Context) {
	handler.handler.ServeHTTP(w, req)
}
