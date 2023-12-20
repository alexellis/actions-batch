package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v57/github"
	names "github.com/inlets/inletsctl/pkg/names"
	"golang.org/x/oauth2"
)

func main() {

	var (
		owner        string
		fileName     string
		tokenFile    string
		privateRepo  bool
		organisation bool
		runsOn       string
		printLogs    bool
	)

	flag.StringVar(&owner, "owner", "actuated-samples", "The owner of the GitHub repository")
	flag.StringVar(&fileName, "file", "", "The name of the file to run via a GitHub Action")
	flag.StringVar(&tokenFile, "token-file", "", "The name of the PAT token file")
	flag.BoolVar(&organisation, "org", true, "Create the repository in an organization")
	flag.StringVar(&runsOn, "runs-on", "actuated", "Runner label for the GitHub action, use ubuntu-latest for a hosted runner")
	flag.BoolVar(&privateRepo, "private", false, "Make the repository private")
	flag.BoolVar(&printLogs, "logs", true, "Print the logs from the workflow run")

	flag.Parse()

	if fileName == "" {
		panic("Please provide a file name")
	}

	if _, err := os.Stat(tokenFile); err != nil && os.IsNotExist(err) {
		panic("Please provide a valid token file")
	}

	repoName := names.GetRandomName(5)
	fmt.Printf("Repo: %s/%q\n", owner, repoName)

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

	orgVal := owner
	if !organisation {
		orgVal = ""
	}

	repo, _, err := client.Repositories.Create(ctx, orgVal, &github.Repository{
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

	fmt.Printf("Delete repo at: https://github.com/%s/%s/settings\n", owner, repoName)

	killCh := make(chan os.Signal, 1)
	signal.Notify(killCh, os.Interrupt)

	defer func() {
		fmt.Printf("Deleting repo: %s/%s\n", owner, repoName)
		_, err := client.Repositories.Delete(ctx, owner, repoName)
		if err != nil {
			log.Printf("failed to delete repo: %s", err)
		}
	}()

	go func() {
		<-killCh
		fmt.Printf("Deleting repo: %s/%s\n", owner, repoName)
		_, err := client.Repositories.Delete(ctx, owner, repoName)
		if err != nil {
			log.Printf("failed to delete repo: %s", err)
		}

		os.Exit(0)
	}()

	if printLogs {
		attempts := 120
		wait := 1 * time.Second

		var workflowRuns *github.WorkflowRuns
		for i := 0; i < attempts; i++ {
			log.Printf("Listing workflow runs for: %s/%s [%d/%d]", owner, repoName, i, attempts)

			wfs, resp, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repoName,
				&github.ListWorkflowRunsOptions{
					Status: "completed",
					Branch: "master",
					ListOptions: github.ListOptions{
						PerPage: 100,
					},
				})
			if err != nil {
				log.Printf("failed to get workflow runs: %s", err)
			}

			if resp.StatusCode == http.StatusNotFound || len(wfs.WorkflowRuns) == 0 {
				log.Printf("No workflow runs (%d) found, waiting %s", resp.StatusCode, wait)
				time.Sleep(wait)
				continue
			} else {
				workflowRuns = wfs
				break
			}
		}

		for _, wf := range workflowRuns.WorkflowRuns {
			fmt.Printf("Getting logs for %d\n", wf.GetID())

			logsURL, resp, err := client.Actions.GetWorkflowRunLogs(ctx, owner, repoName, wf.GetID(),
				1)

			log.Printf("Response: %s", resp.Status)
			if err != nil {
				log.Panicf("failed to get workflow logs: %s", err)
			}

			zip, err := getLogs(logsURL)
			if err != nil {
				log.Panicf("failed to get workflow logs: %s", err)
			}

			tmp := os.TempDir()
			tmpFile, err := os.CreateTemp(tmp, "logs-*.zip")
			if err != nil {
				log.Panicf("failed to create temp file: %s", err)
			}

			if _, err := tmpFile.Write(zip); err != nil {
				log.Panicf("failed to write temp file: %s", err)
			}

			stat, err := os.Stat(tmpFile.Name())
			if err != nil {
				log.Panicf("failed to stat temp file: %s", err)
			}

			zipFile, err := os.Open(tmpFile.Name())
			if err != nil {
				log.Panicf("failed to open temp file: %s", err)
			}
			defer zipFile.Close()

			if err := Unzip(zipFile, stat.Size(), tmp, false); err != nil {
				log.Panicf("failed to unzip file: %s", err)
			}

			tmpDir, err := os.ReadDir(tmp)
			if err != nil {
				log.Panicf("failed to read temp dir: %s", err)
			}

			for _, f := range tmpDir {
				if strings.HasSuffix(f.Name(), ".txt") {
					fmt.Printf("Found file: %s\n---------------------------------\n", f.Name())
					data, _ := os.ReadFile(path.Join(tmp, f.Name()))
					fmt.Printf("%s\n", string(data))
				}
			}

		}
	}
}

func getLogs(logsURL *url.URL) ([]byte, error) {
	fmt.Printf("Getting logs from %s\n", logsURL.String())

	req, err := http.NewRequest(http.MethodGet, logsURL.String(), nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "actuated-batch")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()

		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("failed to get logs, %s", string(body))
	}

	return body, nil

}
