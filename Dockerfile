# build stage
FROM golang:1.11 AS build-env

ADD ./src /go/src/app

RUN mkdir /app && \
    cd /go/src/app && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o directory_s3_watcher directory_s3_watcher.go

# final stage
FROM scratch
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /go/src/app/directory_s3_watcher /directory_s3_watcher
CMD ["/directory_s3_watcher"]
