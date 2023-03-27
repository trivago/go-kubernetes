ARG ARCH=amd64
FROM --platform=${ARCH} cgr.dev/chainguard/go:1.20 as builder

ARG ARCH=amd64

WORKDIR /workspace

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${ARCH}

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/${CMD}/*.go ./cmd/${CMD}/
COPY internal ./internal

RUN go build -ldflags="-s -w" -mod=mod
RUN go test -cover -v ./...
