# Directory watcher & S3 uploader

This app watches a directory (specified either by an ENV var or the -path parameter) and uploads all files 
(non-recursively) to an S3 bucket (specified either by an ENV var or the -bucket parameter).

## Configuration

The following environment variables are supported:
* AWS_ACCESS_KEY_ID
* AWS_SECRET_ACCESS_KEY
* AWS_REGION
* AWS_BUCKET
* WATCH_PATH

## Building

`go build src/directory_s3_watcher.go`

## Building the docker image

`docker build .` 