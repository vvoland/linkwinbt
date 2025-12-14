# golang
FROM golang:1.25.5-alpine AS golang

# golangci-lint
FROM golangci/golangci-lint:v2.7.2 AS golangci-lint

# base
FROM golang AS base
WORKDIR /src

# base-with-git
FROM base AS base-with-git
RUN apk add git

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.sum,destination=go.sum \
    --mount=type=bind,source=go.mod,destination=go.mod \
    go mod download

# test
FROM base AS test
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    go test ./...

# vet
FROM base AS vet
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    go vet ./...

# lint
FROM golangci-lint AS lint

RUN --mount=type=cache,target=/cache \
    --mount=type=bind,source=.,rw \
    GOCACHE=/cache/go-build \
    GOMODCACHE=/cache/mod \
    GOLANGCI_LINT_CACHE=/cache/golangci-lint \
    golangci-lint run --color=always ./...

# gomod
FROM base-with-git AS gomod
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    go mod tidy && git diff --exit-code go.mod go.sum
