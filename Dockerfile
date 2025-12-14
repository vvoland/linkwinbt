# build
FROM golang:1.25.5-alpine AS build

WORKDIR /out
WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.sum,destination=go.sum \
    --mount=type=bind,source=go.mod,destination=go.mod \
    go mod download

ARG TARGETOS
ARG TARGETARCH
ARG SOURCE_DATE_EPOCH=0
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
        -trimpath -buildvcs=false -ldflags="-s -w" \
        -o /out ./...

# binary
FROM scratch AS binary
COPY --from=build /out/* /

# final
FROM alpine:latest AS final
RUN apk add chntpw
COPY --from=build /out/linkwinbt /linkwinbt
ENTRYPOINT ["/linkwinbt"]
