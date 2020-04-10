# Build
FROM golang:1.14-alpine3.11 AS base
COPY . /go/src/sort-it

WORKDIR /go/src/sort-it
RUN apk add git gcc

# Go modules
ENV GO111MODULE=on
RUN go mod download

# Compile
RUN go build -a -tags netgo -ldflags '-w' -o /go/bin/sort-it /go/src/sort-it/main.go

# Package
FROM alpine:3.11
COPY --from=base /go/bin/sort-it /sort-it
RUN apk add exiftool
ENTRYPOINT ["/sort-it"]