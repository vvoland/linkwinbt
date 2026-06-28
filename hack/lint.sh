#!/usr/bin/env bash
# SPDX-License-Identifier: BSD-3-Clause

set -euo pipefail

cd "$(dirname "$0")/.."

ERROR=$'\033[1;31m'
INFO=$'\033[1;34m'
GREY=$'\033[90m'
RESET=$'\033[0m'

# Named volumes used as Go/golangci-lint caches; persist between runs.
GO_CACHE_MOUNTS=(
    -v lint-go-build:/cache/go-build
    -v lint-go-mod:/cache/go-mod
    -v lint-golangci:/cache/golangci-lint
    -e GOCACHE=/cache/go-build
    -e GOMODCACHE=/cache/go-mod
    -e GOLANGCI_LINT_CACHE=/cache/golangci-lint
    -e GOFLAGS=-buildvcs=false
    -e HOME=/tmp
)

BUILD_CACHE_ARGS=()
if [[ -n ${GITHUB_ACTIONS:-} ]]; then
    BUILD_CACHE_ARGS=(
        --cache-from "type=gha,scope=lint"
        --cache-to "type=gha,scope=lint,mode=max"
    )
fi

# When running inside GitHub Actions, forward the env var so tools that
# auto-detect (actionlint) and switch on it find it inside the container.
CI_ENV=()
if [[ -n ${GITHUB_ACTIONS:-} ]]; then
    CI_ENV=(-e GITHUB_ACTIONS=true)
fi

run() {
    printf "  %s\$ %s%s\n" "$GREY" "$*" "$RESET"
    "$@"
}

run_silent() {
    printf "  %s\$ %s%s\n" "$GREY" "$*" "$RESET"
    "$@" >/dev/null
}

build() {
    local target=$1
    printf "%s> Building %s%s\n" "$INFO" "$target" "$RESET"
    run_silent docker buildx build -q --load "${BUILD_CACHE_ARGS[@]}" -f lint.Dockerfile --target "$target" -t "lint/$target" .
}

docker_run() {
    run docker run \
        --network none \
        --security-opt no-new-privileges \
        --rm -w /work "$@"
}

lint_yamllint() {
    build yamllint
    printf "%s> Linting yaml files%s\n" "$INFO" "$RESET"
    local format=${GITHUB_ACTIONS:+github}
    docker_run \
        -v ./.github/workflows:/work/.github/workflows:ro \
        -v ./.yamllint.yaml:/work/.yamllint.yaml:ro \
        lint/yamllint -f "${format:-colored}" .github/workflows
}

lint_actionlint() {
    build actionlint
    printf "%s> Linting GHA yaml files%s\n" "$INFO" "$RESET"
    # actionlint auto-detects $GITHUB_ACTIONS and emits ::error workflow commands.
    docker_run "${CI_ENV[@]}" \
        -v ./.github:/work/.github:ro \
        -v ./.git:/work/.git:ro \
        lint/actionlint -color
}

lint_golangci_lint() {
    build golangci-lint
    printf "%s> Linting Go files%s\n" "$INFO" "$RESET"
    docker_run \
        -v ./:/work \
        -v ./vendor:/work/vendor \
        "${GO_CACHE_MOUNTS[@]}" \
        lint/golangci-lint run --color=always ./...
}

lint_modernize() {
    build modernize
    printf "%s> Running modernize analyzer%s\n" "$INFO" "$RESET"
    docker_run \
        -v ./:/work \
        -v ./vendor:/work/vendor \
        "${GO_CACHE_MOUNTS[@]}" \
        lint/modernize ./...
}

fmt_modernize() {
    build modernize
    printf "%s> Applying modernize fixes%s\n" "$INFO" "$RESET"
    docker_run \
        -v ./:/work \
        -v ./vendor:/work/vendor \
        "${GO_CACHE_MOUNTS[@]}" \
        lint/modernize -fix ./...
}

lint_shfmt() {
    build shfmt
    printf "%s> Linting shell scripts%s\n" "$INFO" "$RESET"
    # shfmt has no --color flag; rely on its isatty(stdout) check via docker -t.
    local tty=()
    [[ -t 1 ]] && tty=(-t)
    docker_run "${tty[@]}" \
        -v ./hack:/work/hack:ro \
        -v ./.editorconfig:/work/.editorconfig:ro \
        lint/shfmt -d hack
}

lint_license() {
    printf "%s> Checking license headers%s\n" "$INFO" "$RESET"

    local missing=()
    local file
    while IFS= read -r file; do
        case "$file" in
            vendor/*) continue ;;
            *.go | *.sh) ;;
            *) continue ;;
        esac

        if ! head -n 5 "$file" | grep -Eq 'SPDX-License-Identifier:'; then
            missing+=("$file")
        fi
    done < <(git ls-files)

    if ((${#missing[@]} > 0)); then
        printf '%s! Missing SPDX license header in:%s\n' "$ERROR" "$RESET" >&2
        printf '  %s\n' "${missing[@]}" >&2
        return 1
    fi

    printf 'All checked files contain an SPDX license header.\n'
}

fmt_shfmt() {
    build shfmt
    printf "%s> Formatting shell scripts%s\n" "$INFO" "$RESET"
    docker_run \
        -v ./hack:/work/hack \
        -v ./.editorconfig:/work/.editorconfig:ro \
        lint/shfmt -w hack
}

fmt_golangci_lint() {
    build golangci-lint
    printf "%s> Formatting Go files%s\n" "$INFO" "$RESET"
    docker_run \
        -v ./:/work \
        -v ./vendor:/work/vendor \
        "${GO_CACHE_MOUNTS[@]}" \
        lint/golangci-lint run --fix ./...
}

case "${1:-}" in
    yamllint) lint_yamllint ;;
    actionlint) lint_actionlint ;;
    golangci-lint) lint_golangci_lint ;;
    modernize) lint_modernize ;;
    shfmt) lint_shfmt ;;
    license) lint_license ;;
    fmt-shfmt) fmt_shfmt ;;
    fmt-golangci-lint) fmt_golangci_lint ;;
    fmt-modernize) fmt_modernize ;;
    *)
        echo "usage: $0 {yamllint|actionlint|golangci-lint|modernize|shfmt|license|fmt-shfmt|fmt-golangci-lint|fmt-modernize}" >&2
        exit 2
        ;;
esac
