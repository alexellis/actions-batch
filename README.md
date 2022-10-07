# actions-batch

A prototype for turning GitHub Actions into a batch job runner.

## Goal

Run a shell script in an isolated, immutable environment.

This works well with actuated or GitHub's hosted runners.

## How it works

1. You write a bash script like the ones in [examples](examples) and pass it in as an argument
1. A new repo is created with a random name in the specified organisation
2. A workflow file is written to the repo along with the shell script, the workflow's only job is to run the shell script and exit
3. The workflow is triggered and you can check the results

## What's left

The part that's left is:

1. Getting a webhook event when the "batch job" is done
2. Collecting the results from the workflow run, perhaps you can post these to S3 directly from bash?
3. Deleting the repo to clean things up

The Rate Limit for a Personal Access Token is quite limited, so this would need to be run as a GitHub App.

## License

MIT