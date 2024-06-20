package templates

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"text/template"
)

//go:embed workflow.yaml
var defaultWorkflow string

func Render(p RenderParams) (string, error) {
	var workflowT *template.Template
	relativeTmplPath := "./workflow.yaml"

	if _, err := os.Stat(relativeTmplPath); err == nil {
		tmpl := template.Must(template.ParseFiles(relativeTmplPath))
		workflowT, err = tmpl.ParseFiles(relativeTmplPath)

		if err != nil {
			log.Panicf("failed to parse workflow template: %s", err)
		}
	} else {
		workflowT = template.Must(template.New("workflow").Parse(defaultWorkflow))
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
