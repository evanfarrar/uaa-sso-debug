package main

import "github.com/evanfarrar/uaa-sso-debug/application"

func main() {
	app := application.NewApplication()
	defer app.Crash()
	app.StartServer()
}
