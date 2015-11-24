package scaling

import (
	"bytes"
	html "html/template"
	"io/ioutil"
	text "text/template"

	"github.com/evanfarrar/uaa-sso-debug/config"
)

type TemplateContext struct {
	AppName                      string
	FromInstanceCount            int
	ToInstanceCount              int
	CPUUtilization               int
	CPUMaxThreshold              int
	CPUMinThreshold              int
	MaxInstanceCount             int
	MinInstanceCount             int
	PlanDuration                 string
	ReadingExpectedInstanceCount int
	BindingExpectedInstanceCount int
}

type TemplateRenderer struct {
	templatesPath string
}

func NewTemplateRenderer() TemplateRenderer {
	env := config.NewEnvironment()
	return TemplateRenderer{
		templatesPath: env.Root + "/workers/scaling/templates/",
	}
}

func (renderer TemplateRenderer) RenderText(name string, context TemplateContext) string {
	tmpl, err := ioutil.ReadFile(renderer.templatesPath + name)
	if err != nil {
		panic(err)
	}

	buffer := bytes.NewBuffer([]byte{})
	template, err := text.New(name).Parse(string(tmpl))
	if err != nil {
		panic(err)
	}

	template.Execute(buffer, context)

	return buffer.String()
}

func (renderer TemplateRenderer) RenderHTML(name string, context TemplateContext) string {
	tmpl, err := ioutil.ReadFile(renderer.templatesPath + name)
	if err != nil {
		panic(err)
	}

	buffer := bytes.NewBuffer([]byte{})
	template, err := html.New(name).Parse(string(tmpl))
	if err != nil {
		panic(err)
	}

	template.Execute(buffer, context)

	return buffer.String()
}
