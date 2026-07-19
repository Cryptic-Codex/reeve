BINARY := reeve
CMD := ./cmd/reeve
BIN := bin/$(BINARY)

.PHONY: build run test vet fmt install clean

build: ## Build the reeve binary into bin/
	go build -o $(BIN) $(CMD)

run: ## Run reeve; pass args with ARGS, e.g. make run ARGS="char elf"
	go run $(CMD) $(ARGS)

test: ## Run all tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all packages
	go fmt ./...

install: ## Install reeve onto your PATH
	go install $(CMD)

clean: ## Remove build artifacts
	rm -rf bin
