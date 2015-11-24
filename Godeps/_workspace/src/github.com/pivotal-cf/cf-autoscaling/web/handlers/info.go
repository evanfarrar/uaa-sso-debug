package handlers

import (
	"net/http"

	"github.com/ryanmoran/stack"
)

type Info struct{}

func NewInfo() Info {
	return Info{}
}

func (handler Info) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
