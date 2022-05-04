# Build
FROM golang:1.18.1-alpine3.15 AS base
RUN apk add alpine-sdk
COPY . /go/src/github.com/aschbacd/sort-it
WORKDIR /go/src/github.com/aschbacd/sort-it
RUN go build -a -tags netgo -ldflags '-w' -o /go/bin/sort-it /go/src/github.com/aschbacd/sort-it

# Package
FROM alpine:3.15.4
COPY --from=base /go/bin/sort-it /sort-it
RUN apk add exiftool
ENTRYPOINT ["/sort-it"]
