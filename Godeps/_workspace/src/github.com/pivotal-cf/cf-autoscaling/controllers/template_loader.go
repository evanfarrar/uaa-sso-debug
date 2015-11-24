package controllers

import (
	"html/template"
	"io/ioutil"

	"github.com/evanfarrar/uaa-sso-debug/config"
)

func LoadTemplate(name string) *template.Template {
	env := config.NewEnvironment()
	source, err := ioutil.ReadFile(env.PublicPath + "templates/" + name + ".html")
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New(name).Parse(string(source))
	if err != nil {
		panic(err)
	}

	return tmpl
}

func LoadFile(name string) *template.Template {
	env := config.NewEnvironment()
	filepath := env.PublicPath + name + ".html"
	tmpl, err := template.ParseFiles(filepath)
	if err != nil {
		panic(err)
	}

	return tmpl
}
