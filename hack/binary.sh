#!/bin/sh
# SPDX-License-Identifier: BSD-3-Clause

: "${OUTDIR:=_build}"

mkdir -p "$OUTDIR"
CGO_ENABLED=0 go build \
    -trimpath -buildvcs=false -ldflags="-s -w" \
    -o "$OUTDIR/linkwinbt" ./cmd/linkwinbt
