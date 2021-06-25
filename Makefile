ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)

GOTEST_MIN = go test -v -timeout 30s
GOTEST = $(GOTEST_MIN) -race
GOTEST_WITH_COVERAGE = $(GOTEST) -coverprofile=coverage.txt -covermode=atomic

.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build-service
build-services: ## Build services Docker images
	$(info Make: Building...)
	docker-compose build

.PHONY: up-services
up-services: ## Run service containers and print logs in stdout
	$(info Make: Starting containers...)
	docker-compose up -d
	@make -s logs

.PHONY: up-env
up-env: ## Run env containers
	$(info Make: Starting containers...)
	docker-compose -f docker-compose-env/docker-compose.yml up -d

.PHONY: down-services
down-services: ## Stop services containers
	$(info Make: Stopping containers...)
	docker-compose down

.PHONY: down-env
down-env: ## Stop env containers
	$(info Make: Stopping containers...)
	docker-compose -f docker-compose-env/docker-compose.yml down --remove-orphans

.PHONY: logs
logs: ## Print services logs in stdout
	docker-compose logs -f keygen links redirects

.PHONY: test-with-coverage
test-with-coverage: ## Run tests with coverage
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  (cd "$${dir}" && \
	    $(GOTEST_WITH_COVERAGE) ./... && \
	    go tool cover -html=coverage.txt -o coverage.html); \
	done

.PHONY: test
test: ## Run tests
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  (cd "$${dir}" && \
	    $(GOTEST) ./...); \
	done

.PHONY: lint
lint: ## Run linter
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  (cd "$${dir}" && \
	    golangci-lint run -E goimports); \
	done

.PHONY: tidy
tidy: ## Run tidy
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  (cd "$${dir}" && \
	    go mod tidy); \
	done

.PHONY: ci
ci: test lint tidy