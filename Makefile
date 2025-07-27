NAME ?= import-visualiser
BUILD_DIR ?= build
OUT_PATH=$(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH)
GIT_RELEASE ?= $(shell git rev-parse --short HEAD)

.PHONY: explain
explain:
	### Welcome
	#
	# This Makefile is used to manage the build and development of the Import Visualiser application.
	#
	### Installation
	#
	# $$ make all
	#
	### Targets
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: ## Install the local dependencies
	go get ./...

.PHONY: lint
lint: ## Vet the code
	golangci-lint run

.PHONY: build
build: ## Build the application
	go build -o $(NAME) .

.PHONY: static
static: ## Build the application
	CGO_ENABLED=0 go build \
		-ldflags "-extldflags -static" \
		-o $(NAME) .

.PHONY: test
test: ## Run the unit tests
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: test-cov
test-cov: test ## Run the unit tests with coverage
	go tool cover -html=coverage.out -o coverage.html
