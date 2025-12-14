# build
FROM golang:1.25.5-alpine AS golang

# golangci-lint
FROM golang AS golangci-lint
RUN apk add git
WORKDIR /go/src
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64

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
FROM base AS lint
COPY --link --from=golangci-lint /go/bin/golangci-lint /usr/bin/golangci-lint

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    golangci-lint run ./...

# gomod
FROM base-with-git AS gomod
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    go mod tidy && git diff --exit-code go.mod go.sum
