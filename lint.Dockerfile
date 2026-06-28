# syntax=docker/dockerfile:1

ARG GOLANG_IMAGE=golang:1.26-alpine
ARG PYTHON_IMAGE=python:3.14-alpine
ARG DISTROLESS_STATIC_IMAGE=gcr.io/distroless/static-debian13:nonroot
ARG UID=1000

FROM ${GOLANG_IMAGE} AS golang-base
RUN --mount=type=cache,id=apk,target=/var/cache/apk,sharing=locked \
    --mount=type=bind,source=hack/install-deps.sh,target=/tmp/install-deps.sh \
    apk add git && sh /tmp/install-deps.sh dev

FROM golang-base AS build-actionlint
ARG ACTIONLINT_VERSION=1.7.12
ENV CGO_ENABLED=0 GOFLAGS=-trimpath
RUN --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    go install "github.com/rhysd/actionlint/cmd/actionlint@v${ACTIONLINT_VERSION}"

FROM ${DISTROLESS_STATIC_IMAGE} AS actionlint
COPY --from=build-actionlint /go/bin/actionlint /usr/local/bin/actionlint
USER ${UID}
WORKDIR /work
ENTRYPOINT ["/usr/local/bin/actionlint"]

FROM golang-base AS build-golangci-lint
ARG GOLANGCI_LINT_VERSION=2.6.2
ENV CGO_ENABLED=0 GOFLAGS=-trimpath
RUN --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v${GOLANGCI_LINT_VERSION}"

FROM golang-base AS build-modernize
# modernize lives inside gopls; the version arg is a gopls release tag.
ARG MODERNIZE_VERSION=v0.21.1
ENV CGO_ENABLED=0 GOFLAGS=-trimpath
RUN --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    go install "golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@${MODERNIZE_VERSION}"

FROM golang-base AS golang-tools
COPY --from=build-golangci-lint /go/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --from=build-modernize /go/bin/modernize /usr/local/bin/modernize
USER ${UID}
WORKDIR /cache/go-build
WORKDIR /cache/go-mod
WORKDIR /cache/golangci-lint
WORKDIR /work
ENTRYPOINT ["/bin/sh"]

FROM golang-tools AS golangci-lint
ENTRYPOINT ["/usr/local/bin/golangci-lint"]

FROM golang-tools AS modernize
ENTRYPOINT ["/usr/local/bin/modernize"]

FROM golang-base AS build-shfmt
ARG SHFMT_VERSION=3.10.0
ENV CGO_ENABLED=0 GOFLAGS=-trimpath
RUN --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    go install "mvdan.cc/sh/v3/cmd/shfmt@v${SHFMT_VERSION}"

FROM ${DISTROLESS_STATIC_IMAGE} AS shfmt
COPY --from=build-shfmt /go/bin/shfmt /usr/local/bin/shfmt
USER ${UID}
WORKDIR /work
ENTRYPOINT ["/usr/local/bin/shfmt"]

FROM ${PYTHON_IMAGE} AS yamllint
ARG YAMLLINT_VERSION=1.37.1
RUN --mount=type=cache,id=pip,target=/root/.cache/pip \
    pip install --root-user-action=ignore "yamllint==${YAMLLINT_VERSION}"
USER ${UID}
WORKDIR /tmp/.cache
WORKDIR /work
ENTRYPOINT ["yamllint"]
