set shell := ["/usr/bin/env", "bash", "-euo", "pipefail", "-c"]

# ------------------------------------------------------------------------------

_default:
  @just -l

build:
  docker build -t "go-kubernetes:latest" .
