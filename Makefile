LDFLAGS=-ldflags "-w -s "

.PHONY: all
all: run

.PHONY: clean
clean: ## Clean bin dir
	rm -rf bin/

.PHONY: fmt
fmt: ## Run go fmt and goimports
	go fmt ./...
	go mod tidy
	go mod download

.PHONY: build
build: ## Build app
	CGO_ENABLED=0 go build ${LDFLAGS} -o ./bin/semaphore ./cmd/semaphore

.PHONY: run
run: build ## Build and run with default config
	bin/semaphore --config cmd/configs/semaphore.config.yml

.PHONY: build-proff ## Build app for proff
build_proff: fmt
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/semaphore cmd/semaphore

.PHONY: build-race
build-race: ## Build app with race flag
	CGO_ENABLED=1 go build -race -o bin/semaphore-race cmd/semaphore

.PHONY: run-race
run-race: build-race ## Build with race flag and run with default config
	bin/semaphore-race --config=configs/semaphore.config.yml

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run -v ./...

.PHONY: gen  ## check and regenerate mock
gen:
	go generate ./...

.PHONY: test
test: fmt gen ## Run tests
	# go test -race -cover -count=1 -short -v ./...
	@go test ./internal...

.PHONY: cover
cover: ## Run tests with cover
	@go test -coverprofile=coverage.out ./internal...

.PHONY: coverr
coverr: cover ## Run tests with cover and open report
	@go tool cover -html=coverage.out

.PHONY: help
help: ## List of commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

