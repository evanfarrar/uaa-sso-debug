package utilities

import "runtime/debug"

func Recover(ident Identifiable) {
	err := recover()
	if err != nil {
		Log(ident, "PANIC: %+v\n", err)
		Log(ident, "%s", debug.Stack())
	}
}
