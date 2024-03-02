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
	"text/tabwriter"
	"time"

	gounits "github.com/docker/go-units"

	"github.com/google/go-github/v57/github"
	"github.com/inlets/inletsctl/pkg/names"
	"golang.org/x/oauth2"

	"github.com/alexellis/actions-batch/pkg"
	"github.com/alexellis/actions-batch/templates"
)

const branch = "master"
const quietUnzip = true

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {

	var (
		owner                string
		fileName             string
		additionalFiles      arrayFlags
		tokenFile            string
		privateRepo          bool
		organisation         bool
		runsOn               string
		deleteRepo           bool
		printLogs            bool
		secretsFrom          string
		maxFetchLogsAttempts int
		fetchLogsInterval    time.Duration
		verbose              bool
		artifactsPath        string
	)

	flag.StringVar(&owner, "owner", "actuated-samples", "The owner of the GitHub repository")
	flag.StringVar(&fileName, "file", "", "The name of the file to run via a GitHub Action")
	flag.Var(&additionalFiles, "additional-file", "Additional files to include in the repository, i.e. --additional-file=Dockerfile --additional-file=Makefile")
	flag.StringVar(&tokenFile, "token-file", "", "The name of the PAT token file")
	flag.BoolVar(&organisation, "org", true, "Create the repository in an organization")
	flag.StringVar(&runsOn, "runs-on", "ubuntu-latest", "Runner label for the GitHub action, use ubuntu-latest for a hosted runner")
	flag.BoolVar(&privateRepo, "private", false, "Make the repository private")
	flag.BoolVar(&printLogs, "logs", true, "Print the logs from the workflow run")
	flag.IntVar(&maxFetchLogsAttempts, "max-attempts", 360, "Maximum number of attempts to fetch logs, this corresponds to job run time so each attempt has a 1 second delay between checking")
	flag.DurationVar(&fetchLogsInterval, "interval", 1*time.Second, "Interval between checking for logs")
	flag.StringVar(&secretsFrom, "secrets-from", "", "Create secrets from the files on disk, converting i.e. openfaas-password to: OPENFAAS_PASSWORD, and making that available via an environment variable.")
	flag.BoolVar(&deleteRepo, "delete", true, "Delete the repository after the run")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging")
	flag.StringVar(&artifactsPath, "out", "", "Path to use to unzip the artifacts folder from the build, if there is one")

	flag.Parse()

	fmt.Printf(

		`┏━┓┏━╸╺┳╸╻┏━┓┏┓╻┏━┓   ┏┓ ┏━┓╺┳╸┏━╸╻ ╻
┣━┫┃   ┃ ┃┃ ┃┃┗┫┗━┓╺━╸┣┻┓┣━┫ ┃ ┃  ┣━┫
╹ ╹┗━╸ ╹ ╹┗━┛╹ ╹┗━┛   ┗━┛╹ ╹ ╹ ┗━╸╹ ╹
By Alex Ellis %d - %s (%s)

`, time.Now().Year(), pkg.Version, pkg.GitCommit)

	if fileName == "" {
		panic("--file is required")
	}

	if _, err := os.Stat(tokenFile); err != nil && os.IsNotExist(err) {
		panic("Please provide a valid token file")
	}

	if len(secretsFrom) > 0 {
		if stat, err := os.Stat(secretsFrom); err != nil && os.IsNotExist(err) {
			panic("Please provide a valid folder for the secrets-from flag")
		} else if !stat.IsDir() {
			panic(fmt.Sprintf("%s is not a folder", secretsFrom))
		}
	}

	repoName := names.GetRandomName(5)
	fmt.Printf("Job file: %s\n", path.Base(fileName))
	fmt.Printf("Additional files: %s\n", path.Base(strings.Join(additionalFiles, ", ")))
	fmt.Printf("Repo: https://github.com/%s/%s\n", owner, repoName)

	t := os.TempDir()

	tmp, err := os.MkdirTemp(t, repoName)
	if err != nil {
		log.Panicf("failed to create temp dir %s, %s", t, err)
	}

	defer os.RemoveAll(tmp)

	if verbose {
		fmt.Printf("Writing templates to: %s\n", tmp)
	}

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

	if _, _, err := client.Repositories.Create(ctx, orgVal, &github.Repository{
		Name:          github.String(repoName),
		Private:       github.Bool(privateRepo),
		DefaultBranch: github.String(branch),
	}); err != nil {
		log.Panicf("failed to create repo: %s", err)
	}

	if deleteRepo {
		defer func() {
			fmt.Printf("Deleting repo: %s/%s\n", owner, repoName)
			_, err := client.Repositories.Delete(ctx, owner, repoName)
			if err != nil {
				log.Panicf("failed to delete repo: %s", err)
			}
		}()
	} else {
		fmt.Printf("Delete repo at: https://github.com/%s/%s/settings\n", owner, repoName)
	}

	secretsMap := map[string]string{}
	if len(secretsFrom) > 0 {
		if mapped, err := createSecrets(ctx, client, owner, repoName, secretsFrom); err != nil {
			log.Panicf("failed to create secrets: %s", err)
		} else {
			secretsMap = mapped
		}
	}

	// job.sh
	out, err := templates.Render(templates.RenderParams{
		Name:    repoName,
		Login:   login,
		Date:    time.Now().String(),
		RunsOn:  runsOn,
		Secrets: secretsMap,
	})
	if err != nil {
		log.Panicf("failed to render workflow template: %s", err)
	}

	if _, err := f.WriteString(out); err != nil {
		log.Panicf("failed to write workflow file: %s", err)
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

	fileBytes, err := os.ReadFile(jobFile)
	if err != nil {
		log.Panicf("failed to read job file: %s", err)
	}
	if _, _, err := client.Repositories.CreateFile(ctx, owner, repoName, "job.sh",
		&github.RepositoryContentFileOptions{
			Message: github.String("Add job.sh"),
			Content: []byte(fileBytes),
			Author: &github.CommitAuthor{
				Name:  github.String("actuated-batch"),
				Email: github.String("actuated-samples@users.noreply.github.com"),
			},
			Branch: github.String(branch),
		}); err != nil {
		log.Panicf("failed to create workflow file: %s", err)
	}

	// Additional files
	for _, afile := range additionalFiles {
		if _, err := os.Stat(afile); os.IsNotExist(err) {
			log.Panicf("ERROR: %s does not exist, Error: %s", afile, err)
		}
		fileBytes, err := os.ReadFile(afile)
		if err != nil {
			log.Panicf("failed to read file: %s", err)
		}

		if _, _, err := client.Repositories.CreateFile(ctx, owner, repoName, path.Base(afile),
			&github.RepositoryContentFileOptions{
				Message: github.String(fmt.Sprintf("Add %s", path.Base(afile))),
				Content: fileBytes,
				Author: &github.CommitAuthor{
					Name:  github.String("actuated-batch"),
					Email: github.String("actuated-samples@users.noreply.github.com"),
				},
				Branch: github.String(branch),
			}); err != nil {
			log.Panicf("failed to create file: %s", err)
		}
	}

	// GitHub Actions workflow upload
	fileBytes, err = os.ReadFile(actionsFile)
	if err != nil {
		log.Panicf("failed to read workflow file: %s", err)
	}

	if _, _, err = client.Repositories.CreateFile(ctx, owner, repoName, ".github/workflows/workflow.yaml", &github.RepositoryContentFileOptions{
		Message: github.String("Add workflow for GitHub Actions"),
		Content: fileBytes,
		Author: &github.CommitAuthor{
			Name:  github.String("actuated-batch"),
			Email: github.String("actuated-samples@users.noreply.github.com"),
		},
		Branch: github.String(branch),
	}); err != nil {
		log.Panicf("failed to create workflow file: %s", err)
	}

	st := time.Now()

	killCh := make(chan os.Signal, 1)
	signal.Notify(killCh, os.Interrupt)

	go func() {
		<-killCh
		fmt.Printf("Deleting repo: %s/%s\n", owner, repoName)
		_, err := client.Repositories.Delete(ctx, owner, repoName)
		if err != nil {
			log.Printf("failed to delete repo: %s", err)
		}

		os.Exit(0)
	}()

	fmt.Printf("----------------------------------------\n")
	fmt.Printf("View job at: \nhttps://github.com/%s/%s/actions\n", owner, repoName)
	fmt.Printf("----------------------------------------\n")

	if printLogs {
		var runStart time.Time
		var runEnd time.Time

		wait := fetchLogsInterval

		var workflowRuns *github.WorkflowRuns
		fmt.Printf("Listing workflow runs for: %s/%s max attempts: %d (interval: %s)\n",
			owner, repoName, maxFetchLogsAttempts, fetchLogsInterval.Round(time.Second))

		for i := 0; i < maxFetchLogsAttempts; i++ {

			wfs, resp, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repoName,
				&github.ListWorkflowRunsOptions{
					Status: "completed",
					Branch: branch,
					ListOptions: github.ListOptions{
						PerPage: 100,
					},
				})
			if err != nil {
				log.Printf("failed to get workflow runs: %s", err)
			}

			if len(wfs.WorkflowRuns) > 0 {
				runStart = wfs.WorkflowRuns[0].GetRunStartedAt().Time
				runEnd = wfs.WorkflowRuns[0].GetUpdatedAt().Time
			}

			if resp.StatusCode == http.StatusNotFound || len(wfs.WorkflowRuns) == 0 {
				// log.Printf("No workflow runs (%d) found, waiting %s", resp.StatusCode, wait)
				fmt.Fprintf(os.Stderr, ".")
				time.Sleep(wait)
				continue
			} else {
				fmt.Fprintf(os.Stderr, "..OK\n")

				workflowRuns = wfs
				break
			}
		}

		done := time.Now()
		for _, wf := range workflowRuns.WorkflowRuns {
			fmt.Printf("Getting logs for %d\n", wf.GetID())

			const maxRedirects = 1
			logsURL, resp, err := client.Actions.GetWorkflowRunLogs(ctx,
				owner,
				repoName,
				wf.GetID(),
				maxRedirects)

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

			tmpLogs, err := os.MkdirTemp(tmp, "logs-*")
			if err != nil {
				log.Panicf("failed to create temp dir under: %s, %s", tmp, err)
			}

			log.Printf("Unzipping to: %s", tmpLogs)

			defer os.RemoveAll(tmpLogs)

			if err := Unzip(zipFile, stat.Size(), tmpLogs, quietUnzip); err != nil {
				log.Panicf("failed to unzip file: %s", err)
			}

			tmpLogsDir, err := os.ReadDir(tmpLogs)
			if err != nil {
				log.Panicf("failed to read temp dir: %s", err)
			}

			ignoreLogs := []string{
				"Upload Artifact",
				"Complete job",
				"Check for uploads",
			}
			for _, f := range tmpLogsDir {
				if strings.HasSuffix(f.Name(), ".txt") {

					ignoreLog := false
					for _, ignore := range ignoreLogs {
						if strings.Contains(f.Name(), ignore) {
							ignoreLog = true
							break
						}
					}

					// So that we see the program output
					if !ignoreLog {
						fmt.Printf("Found file: %s\n---------------------------------\n", f.Name())
						data, err := os.ReadFile(path.Join(tmpLogs, f.Name()))
						if err != nil {
							log.Panicf("failed to read file: %s", err)
						}

						fmt.Printf("%s\n", string(data))
					}
				}
			}

			if err := downloadArtifacts(ctx, client, owner, repoName, wf.GetID(), artifactsPath); err != nil {
				log.Printf("failed to download artifacts: %s", err)
			}
		}

		t := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		// Queued time | Job duration | Total time

		fmt.Fprintf(t, "QUEUED\tDURATION\tTOTAL\n")
		fmt.Fprintf(t, "%s\t%s\t%s\n", runStart.Sub(st).Round(time.Second), runEnd.Sub(runStart).Round(time.Second), done.Sub(st).Round(time.Second))
		fmt.Fprintf(t, "\n")
		t.Flush()
	}
}

func downloadArtifacts(ctx context.Context, client *github.Client, owner, repoName string, wfID int64, artifactsPath string) error {
	artifacts, _, err := client.Actions.ListWorkflowRunArtifacts(ctx, owner, repoName, wfID, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return err
	}

	for _, a := range artifacts.Artifacts {
		dlUrl, dlUrlRes, err := client.Actions.DownloadArtifact(ctx, owner, repoName, a.GetID(), 1)
		if err != nil {
			return err
		}

		if dlUrlRes.Body != nil {
			defer dlUrlRes.Body.Close()
		}

		if dlUrlRes.StatusCode != http.StatusOK && dlUrlRes.StatusCode != http.StatusFound {
			return fmt.Errorf("failed to get download URL with status: %d", dlUrlRes.StatusCode)
		}

		req, err := http.NewRequest(http.MethodGet, dlUrl.String(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "actuated-batch")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)

			return fmt.Errorf("failed to get logs with status: %d, body: %s", res.StatusCode, string(body))
		}

		tmp := os.TempDir()
		tmpFile, err := os.CreateTemp(tmp, a.GetName())
		if err != nil {
			return err
		}

		if _, err := io.Copy(tmpFile, res.Body); err != nil {
			return err
		}

		outPath := ""
		if len(artifactsPath) > 0 {
			outPath = artifactsPath
		}

		artifactsPath, err := unzipArtifacts(tmpFile.Name(), outPath)
		if err != nil {
			return err
		}

		artifactsDir, err := os.ReadDir(artifactsPath)
		if err != nil {
			return err
		}

		fmt.Printf("Contents of: %s\n\n", artifactsPath)
		t := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(t, "FILE\tSIZE\n")
		for _, f := range artifactsDir {
			i, _ := f.Info()
			i.Size()
			fmt.Fprintf(t, "%s\t%s\n", f.Name(), gounits.HumanSize(float64(i.Size())))
		}

		fmt.Fprintf(t, "\n")
		t.Flush()

	}

	return nil
}

func unzipArtifacts(target, outPath string) (string, error) {

	targetPath := ""
	if len(outPath) > 0 {
		targetPath = outPath
	} else {
		tmp := os.TempDir()
		tmpPath, err := os.MkdirTemp(tmp, "artifacts-*")
		if err != nil {
			return "", fmt.Errorf("failed to create temp dir %s, %w", tmp, err)
		}
		targetPath = tmpPath
	}

	f, err := os.Open(target)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	if err := Unzip(f, stat.Size(), targetPath, quietUnzip); err != nil {
		return "", fmt.Errorf("failed to unzip file: %w", err)
	}

	return targetPath, nil
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
