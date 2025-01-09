ARG ARCH=amd64

FROM --platform=${ARCH} cgr.dev/chainguard/go:latest AS builder

ARG ARCH=amd64
ARG CMD=

WORKDIR /workspace

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${ARCH}
ENV CMD=${CMD}

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY cmd/${CMD}/*.go ./cmd/${CMD}/
#COPY internal ./internal

RUN go test -cover -v ./...
RUN go build -mod=mod ./cmd/${CMD}
