VERSION_YAMLLINT := 1.37.1
VERSION_ACTIONLINT := 1.7.9

RUN := docker run --network none --security-opt no-new-privileges -u 1000:1000 --rm -w /work

COLOR_INFO := \033[1;34m
COLOR_RESET := \033[0m

.PHONY: lint
lint: lint-yaml

.PHONY: lint-yaml
lint-yaml: lint-yaml-style lint-yaml-gha

.PHONY: lint-yaml-style
lint-yaml-style:
	@printf "$(COLOR_INFO)> %s$(COLOR_RESET)\n" "Linting yaml files"
	$(RUN) \
		-v "./.github/workflows:/work/.github/workflows:ro" \
		-v "./.yamllint.yaml:/work/.yamllint.yaml" \
		giantswarm/yamllint:$(VERSION_YAMLLINT) \
		-f $(if $(GITHUB_WORKFLOW),github,colored) \
		.github/workflows

.PHONY: lint-yaml-gha
lint-yaml-gha:
	@printf "$(COLOR_INFO)> %s$(COLOR_RESET)\n" "Linting GHA yaml files"
	$(RUN) \
		-v "./.yamllint.yml:/work/.yamllint.yml:ro" \
		-v "./.github:/work/.github:ro" \
		-v "./.git:/work/.git:ro" \
		rhysd/actionlint:$(VERSION_ACTIONLINT) -color
