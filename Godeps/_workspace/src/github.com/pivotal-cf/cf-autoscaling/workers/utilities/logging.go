package utilities

import (
	"github.com/evanfarrar/uaa-sso-debug/log"
)

type Identifiable interface {
	Identifier() string
}

func Log(ident Identifiable, format string, v ...interface{}) {
	log.Printf(ident.Identifier()+" "+format, v...)
}
