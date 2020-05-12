SHELL := /usr/bin/env bash -o errexit -o nounset -o pipefail
.PHONY: *

# thanks to https://gist.github.com/mpneuried/0594963ad38e68917ef189b4e6a269db
# ENV_FILE ?= _build/local.env
# include $(ENV_FILE)


# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
            {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## build binary
	go build -o bin/gcs-proxy main/main.go
	@printf "\033[36m%-30s\033[0m %s\n" "'$@' finished successfully!"

dev-run: ## run app with with development defaults
	./bin/gcs-proxy config.toml
	@printf "\033[36m%-30s\033[0m %s\n" "'$@' finished successfully!"

docker-build: ## build docker image
	docker build -t afiore/gcs-proxy:latest .
	@printf "\033[36m%-30s\033[0m %s\n" "'$@' finished successfully!"

dev-docker-run: ## run docker image with development defaults
	docker run --rm --volume $(CURDIR):/tmp afiore/gcs-proxy:latest /tmp/config.toml
	@printf "\033[36m%-30s\033[0m %s\n" "'$@' finished successfully!"
