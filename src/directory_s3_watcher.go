package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/radovskyb/watcher"
	"log"
	"os"
	"time"
)

func main() {
	var bucket, basePath string

	flag.StringVar(&basePath, "path", os.Getenv("WATCH_PATH"), "path to the directory to watch")
	flag.StringVar(&bucket, "bucket", os.Getenv("AWS_BUCKET"), "the bucket to upload the files to")

	flag.Parse()

	if "" == basePath {
		fmt.Println("No path provided")
		os.Exit(100)
	}

	if "" == bucket {
		fmt.Println("No bucket set")
		os.Exit(100)
	}

	fmt.Printf("Watching %s and uploading all files to the %s bucket\n\n", basePath, bucket)

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	w := watcher.New()

	// Only notify rename and move events.
	w.FilterOps(watcher.Create, watcher.Write)

	go func() {
		for {
			select {
			case event := <-w.Event:
				if !event.FileInfo.IsDir() {
					uploadFile(event.Path, event.FileInfo.Name(), bucket, svc)
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	w.IgnoreHiddenFiles(true)

	// Watch this folder for changes.
	if err := w.Add(basePath); err != nil {
		log.Fatalln(err)
	}

	// Upload all files from the watched directory
	for path, f := range w.WatchedFiles() {
		if !f.IsDir() {
			uploadFile(path, f.Name(), bucket, svc)
		}
	}

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}

func uploadFile(path string, key string, bucket string, svc *s3.S3) {
	fmt.Printf("Uploading %s as %s: ", path, key)
	file, err := os.Open(path)

	if err != nil {
		fmt.Println("Failed to open file", path, err)
		os.Exit(1)
	}

	_, serr := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})

	if serr != nil {
		if aerr, ok := serr.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			// If the SDK can determine the request or retry delay was canceled
			// by a context the CanceledErrorCode error code will be returned.
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", serr)
		} else {
			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", serr)
		}
		os.Exit(1)
	}

	fmt.Println("complete")
}
