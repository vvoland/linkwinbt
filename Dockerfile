# build
FROM golang:1.24-alpine AS build

WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.sum,destination=go.sum \
    --mount=type=bind,source=go.mod,destination=go.mod \
    go mod download

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    CGO_ENABLED=0 GOOS=linux go build -o /linkwin ./cmd/linkwin/main.go

# final
FROM alpine:latest
RUN apk add chntpw
COPY --from=build /linkwin /linkwin
ENTRYPOINT ["/linkwin"]
