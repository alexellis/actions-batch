package templates

import (
	"bytes"
	"log"
	"text/template"
)

func Render(p RenderParams) (string, error) {
	tmpl := template.Must(template.ParseFiles("./templates/workflow.yaml"))
	workflowT, err := tmpl.ParseFiles("./templates/workflow.yaml")
	if err != nil {
		log.Panicf("failed to parse workflow template: %s", err)
	}

	buf := bytes.NewBuffer(nil)

	if err := workflowT.Execute(buf, p); err != nil {
		log.Panicf("failed to execute workflow template: %s", err)
	}

	return buf.String(), nil
}

type RenderParams struct {
	Name    string
	Login   string
	Date    string
	RunsOn  string
	Secrets map[string]string
}
