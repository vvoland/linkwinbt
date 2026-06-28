# build
FROM --platform=$BUILDPLATFORM golang:1.26-trixie AS build

WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=. \
    GOOS=$TARGETOS GOARCH=$TARGETARCH OUTDIR=/out sh ./hack/binary.sh

# binary
FROM scratch AS binary
COPY --from=build /out/* /

# final
FROM alpine:latest AS final
RUN apk add chntpw
COPY --from=build /out/linkwinbt /linkwinbt
ENTRYPOINT ["/linkwinbt"]
