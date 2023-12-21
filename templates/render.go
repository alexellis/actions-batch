package templates

import (
	"bytes"
	"log"
	"os"
	"text/template"
)

func Render(p RenderParams) (string, error) {

	tmplPath := "./templates/workflow.yaml"
	if _, err := os.Stat(tmplPath); err != nil && os.IsNotExist(err) {
		tmplPath = "./workflow.yaml"
	}

	tmpl := template.Must(template.ParseFiles(tmplPath))
	workflowT, err := tmpl.ParseFiles(tmplPath)
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
