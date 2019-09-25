# How To Run Sims In The Cloud

Sims run with AWS batch, with results uploaded to S3

## AWS Batch

In AWS batch you define:

- a "compute environment"--just how many machines you want (and of what kind)
- a "job queue"--just a place to put jobs (pairs them with a compute environment)
- a "job definition"--a template for jobs

Then to run stuff you create "jobs" and submit them to a job queue.

The number of machine running auto-scales to match the number of jobs. When there are no jobs there are no machines, so you don't pay for anything.

Jobs are defined as a docker image (assumed hosted on dockerhub) and a command string.  
>e.g. `kava/kava-sim:version1`, `go test ./app`

This can run sims but doesn't collect the results. This is handled by a custom script.

## Running sims and uploading to S3

The dockerfile in this repo defines the docker image to run sims. It's just a normal app, but with the aws cli included, and the custom script.

The custom script reads some input args, runs a sim and uploads the stdout and stderr to a S3 bucket.

AWS Batch allows for "array jobs" which are a way of specifying many duplicates of a job, each with a different index passed in as an env var.

### Steps

- create and submit a new array job (based of the job definition) with
  - image `kava/kava-sim:<some-version>`
  - command `run-then-upload.sh <starting-seed> <num-blocks> <block-size>`
  - array size of how many sims you want to run
- any changes needed to the code or script necessitates a rebuild:
  - `docker build -f simulations/Dockerfile -t kava/kava-sim:<some-version> .`
  - `docker push kava/kava-sim:<some-version>`

### Tips

- click on the compute environment name, to get details, then click the link ECS Cluster Name to get details on the actual machines running
- for array jobs, click the job name to get details of the individual jobs
