BINARY_NAME=budget-ingest

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test build vendor

all: help

## Build:
build: ## Build your project and put the output binary in out/bin/
	mkdir -p out/bin
	GO111MODULE=on go build -mod vendor -o out/bin/$(BINARY_NAME) .

clean: ## Remove build related file
	rm -fr ./bin
	rm -fr ./out

vendor: ## Copy of all packages needed to support builds and tests in the vendor directory
	go mod vendor

serve: ## Run the go code
	docker compose --profile serve up

serve-local: vendor ## Run the go code changes that are locally made
	docker image rm ghcr.io/carson-networks/budget-ingest:local || true
	docker compose --profile serve-local build
	docker compose --profile serve-local up

serve-local-rebuild: vendor ## Rebuild the local changes
	docker compose up --detach --build budget-ingest-local

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

format: ## Format imports
	find . -type f -name '*.go' ! -path "*/vendor/*" | xargs go tool goimports -local -l -w

generate-mocks: ## Generate mocks
	go tool mockery
