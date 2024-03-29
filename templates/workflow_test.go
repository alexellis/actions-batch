package templates

import (
	"log"
	"os"
	"testing"
)

func Test_Workflow_WithoutSecrets(t *testing.T) {

	res, err := Render(RenderParams{
		Name:    "test",
		Login:   "test",
		Date:    "test",
		RunsOn:  "test",
		Secrets: map[string]string{},
	})

	if err != nil {
		t.Fatalf("failed to render workflow: %s", err)
	}
	want := `name: workflow

# Generated by alexellis/actuated-batch at: test
# Job requested by test

on:
  pull_request:
    branches:
      - '*'
  push:
    branches:
      - master
      - main

permissions:
  id-token: write
  contents: read
  checks: read
  actions: read
  issues: read
  packages: write
  pull-requests: read
  repository-projects: read
  statuses: read

env:
  HAS_UPLOADS: 0

jobs:
  workflow:
    name: test
    runs-on: test
    steps:
      - uses: actions/checkout@v1
      - name: Run the job
        run: |
          echo "HAS_UPLOADS=0" >> $GITHUB_ENV
          chmod +x ./job.sh
          ./job.sh



      - name: "Check for uploads"
        run: if [ -d "./uploads" ]; then echo "HAS_UPLOADS=1" >> $GITHUB_ENV ; fi

      - name: 'Upload Artifact'
        uses: actions/upload-artifact@v4
        if: ${{ env.HAS_UPLOADS == '1'}}
        with:
          if-no-files-found: ignore
          name: uploads
          path: ./uploads/
          retention-days: 1
          compression-level: 1
`
	got := res

	if got != want {
		tmp := os.TempDir()
		tmpFile, _ := os.CreateTemp(tmp, "test-*")

		log.Printf("tmpFile: %s", tmpFile.Name())
		os.WriteFile(tmpFile.Name(), []byte(got), 0644)

		t.Fatalf("want\n%q\n\ngot\n%q\n\n", want, got)
	}
}

func Test_Workflow_WithSecrets(t *testing.T) {

	res, err := Render(RenderParams{
		Name:   "test",
		Login:  "test",
		Date:   "test",
		RunsOn: "test",
		Secrets: map[string]string{
			"OPENFAAS_GATEWAY": "OPENFAAS_GATEWAY",
			"OPENFAAS_URL":     "OPENFAAS_URL",
		},
	})

	if err != nil {
		t.Fatalf("failed to render workflow: %s", err)
	}
	want := `name: workflow

# Generated by alexellis/actuated-batch at: test
# Job requested by test

on:
  pull_request:
    branches:
      - '*'
  push:
    branches:
      - master
      - main

permissions:
  id-token: write
  contents: read
  checks: read
  actions: read
  issues: read
  packages: write
  pull-requests: read
  repository-projects: read
  statuses: read

env:
  HAS_UPLOADS: 0

jobs:
  workflow:
    name: test
    runs-on: test
    steps:
      - uses: actions/checkout@v1
      - name: Run the job
        run: |
          echo "HAS_UPLOADS=0" >> $GITHUB_ENV
          chmod +x ./job.sh
          ./job.sh


        env:
          OPENFAAS_GATEWAY: ${{ secrets.OPENFAAS_GATEWAY }}
          OPENFAAS_URL: ${{ secrets.OPENFAAS_URL }}

      - name: "Check for uploads"
        run: if [ -d "./uploads" ]; then echo "HAS_UPLOADS=1" >> $GITHUB_ENV ; fi

      - name: 'Upload Artifact'
        uses: actions/upload-artifact@v4
        if: ${{ env.HAS_UPLOADS == '1'}}
        with:
          if-no-files-found: ignore
          name: uploads
          path: ./uploads/
          retention-days: 1
          compression-level: 1
`
	got := res

	if got != want {

		tmp := os.TempDir()
		tmpFile, _ := os.CreateTemp(tmp, "test-*")

		log.Printf("tmpFile: %s", tmpFile.Name())
		os.WriteFile(tmpFile.Name(), []byte(got), 0644)

		t.Fatalf("want\n%q\n\ngot\n%q\n\n", want, got)
	}
}
