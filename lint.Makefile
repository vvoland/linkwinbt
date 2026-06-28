.PHONY: lint
lint: yaml go shell license

.PHONY: yaml
yaml: yamllint actionlint

.PHONY: go
go: modernize golangci-lint

.PHONY: shell
shell: shfmt

.PHONY: license
license:
	@hack/lint.sh license

.PHONY: yamllint
yamllint:
	@hack/lint.sh yamllint

.PHONY: actionlint
actionlint:
	@hack/lint.sh actionlint

.PHONY: golangci-lint
golangci-lint:
	@hack/lint.sh golangci-lint

.PHONY: modernize
modernize:
	@hack/lint.sh modernize

.PHONY: shfmt
shfmt:
	@hack/lint.sh shfmt

.PHONY: fmt
fmt: fmt-go fmt-shell

.PHONY: fmt-go
fmt-go: fmt-modernize fmt-golangci-lint

.PHONY: fmt-shell
fmt-shell: fmt-shfmt

.PHONY: fmt-golangci-lint
fmt-golangci-lint:
	@hack/lint.sh fmt-golangci-lint

.PHONY: fmt-modernize
fmt-modernize:
	@hack/lint.sh fmt-modernize

.PHONY: fmt-shfmt
fmt-shfmt:
	@hack/lint.sh fmt-shfmt
