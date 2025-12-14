# build
FROM golang:1.25.5-alpine AS build

WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.sum,destination=go.sum \
    --mount=type=bind,source=go.mod,destination=go.mod \
    go mod download

ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /linkwinbt ./cmd/linkwinbt/main.go

# binary
FROM scratch AS binary
COPY --from=build /linkwinbt /linkwinbt

# final
FROM alpine:latest AS final
RUN apk add chntpw
COPY --from=build /linkwinbt /linkwinbt
ENTRYPOINT ["/linkwinbt"]
