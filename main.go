package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v47/github"
	names "github.com/inlets/inletsctl/pkg/names"
	"golang.org/x/oauth2"
)

func main() {

	var (
		owner       string
		fileName    string
		tokenFile   string
		privateRepo bool
		runsOn      string
	)

	flag.StringVar(&owner, "owner", "actuated-samples", "The owner of the GitHub repository")
	flag.StringVar(&fileName, "file", "", "The name of the file to run via a GitHub Action")
	flag.StringVar(&tokenFile, "token-file", "", "The name of the PAT token file")
	flag.StringVar(&runsOn, "runs-on", "actuated", "Runner label for the GitHub action, use ubuntu-latest for a hosted runner")
	flag.BoolVar(&privateRepo, "private", false, "Make the repository private")
	flag.Parse()

	if fileName == "" {
		panic("Please provide a file name")
	}

	if _, err := os.Stat(tokenFile); err != nil && os.IsNotExist(err) {
		panic("Please provide a valid token file")
	}

	repoName := names.GetRandomName(5)
	fmt.Printf("repoName: %q\n", repoName)

	t := os.TempDir()

	tmp, err := os.MkdirTemp(t, repoName)
	if err != nil {
		log.Panicf("failed to create temp dir %s, %s", t, err)
	}

	// defer os.RemoveAll(tmp)

	tmpl := template.Must(template.ParseFiles("templates/workflow.yaml"))
	workflowT, err := tmpl.ParseFiles("templates/workflow.yaml")
	if err != nil {
		log.Panicf("failed to parse workflow template: %s", err)
	}

	fmt.Printf("tmp: %q\n", tmp)
	os.MkdirAll(path.Join(tmp, ".github/workflows"), os.ModePerm)
	actionsFile := path.Join(tmp, "/.github/workflows/workflow.yaml")
	f, err := os.Create(actionsFile)
	if err != nil {
		log.Panicf("failed to create workflow file: %s", err)
	}
	defer f.Close()

	login := "unknown"
	loginU, _ := user.Current()
	if loginU != nil {
		login = loginU.Username
	}

	if err := workflowT.Execute(f, map[string]string{
		"Name":   repoName,
		"Login":  login,
		"Date":   time.Now().String(),
		"RunsOn": runsOn,
	}); err != nil {
		log.Panicf("failed to execute workflow template: %s", err)
	}

	f.Close()

	fIn, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		log.Panicf("failed to open file: %s", err)
	}
	defer fIn.Close()

	jobFile := path.Join(tmp, "/job.sh")
	fsh, err := os.Create(jobFile)
	if err != nil {
		log.Panicf("failed to create workflow file: %s", err)
	}
	defer fsh.Close()

	if _, err := io.Copy(fsh, fIn); err != nil {
		log.Panicf("failed to copy file: %s", err)
	}

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		log.Panicf("failed to read token file: %s", err)
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(string(token))},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	repo, _, err := client.Repositories.Create(ctx, owner, &github.Repository{
		Name:    github.String(repoName),
		Private: github.Bool(privateRepo),
	})
	if err != nil {
		log.Panicf("failed to create repo: %s", err)
	}

	fmt.Printf("created repo: %s\n", repo.GetHTMLURL())

	fileBytes, err := os.ReadFile(jobFile)
	if err != nil {
		log.Panicf("failed to read job file: %s", err)
	}
	r, _, err := client.Repositories.CreateFile(ctx, owner, repoName, "job.sh",
		&github.RepositoryContentFileOptions{
			Message: github.String("Add job.sh"),
			Content: []byte(fileBytes),
			Author: &github.CommitAuthor{
				Name:  github.String("actuated-batch"),
				Email: github.String("actuated-samples@users.noreply.github.com"),
			},
		})
	if err != nil {
		log.Panicf("failed to create workflow file: %s", err)
	}

	fmt.Printf("Wrote file %s\n", r.GetHTMLURL())

	fileBytes, err = os.ReadFile(actionsFile)
	if err != nil {
		log.Panicf("failed to read workflow file: %s", err)
	}

	r, _, err = client.Repositories.CreateFile(ctx, owner, repoName, ".github/workflows/workflow.yaml", &github.RepositoryContentFileOptions{
		Message: github.String("Add workflow for GitHub Actions"),
		Content: fileBytes,
		Author: &github.CommitAuthor{
			Name:  github.String("actuated-batch"),
			Email: github.String("actuated-samples@users.noreply.github.com"),
		},
	})
	if err != nil {
		log.Panicf("failed to create workflow file: %s", err)
	}

	fmt.Printf("Wrote file %s\n", r.GetHTMLURL())

}
