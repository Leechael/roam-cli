BINARY_NAME := roam-cli
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(BINARY_NAME)
OUT_DIR := dist
CMD := ./cmd/roam-cli
GOFLAGS ?= -buildvcs=false

OP_PLUGIN_NAME := roamresearch
OP_PLUGIN_DIR := contrib/1password-plugin
OP_PLUGIN_BIN := $(BIN_DIR)/$(OP_PLUGIN_NAME)

export GOFLAGS

.PHONY: tidy fmt test bdd-test ci build run install clean cross-build help
.PHONY: op-plugin-test op-plugin-build op-plugin-install-local

help:
	@echo "Targets:"
	@echo "  make tidy         - go mod tidy"
	@echo "  make fmt          - gofmt ./cmd ./internal"
	@echo "  make test         - run unit tests"
	@echo "  make bdd-test     - run BDD tests"
	@echo "  make ci           - fmt check + vet + tests + build"
	@echo "  make build        - build local binary to ./$(BIN_PATH)"
	@echo "  make run          - run CLI (pass ARGS='...')"
	@echo "  make install      - install binary to GOPATH/bin"
	@echo "  make clean        - remove build artifacts"
	@echo "  make cross-build  - build darwin/linux amd64/arm64 binaries"
	@echo "  make op-plugin-test        - run 1Password plugin tests"
	@echo "  make op-plugin-build       - build 1Password plugin binary"
	@echo "  make op-plugin-install-local - build and install plugin locally"


tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal ./tests

test:
	go test ./... -count=1

bdd-test:
	go test -tags=bdd ./tests/bdd/... -count=1

ci:
	@unformatted=$$(gofmt -l ./cmd ./internal ./tests); \
	if [ -n "$$unformatted" ]; then echo "Unformatted files:"; echo "$$unformatted"; exit 1; fi
	go vet ./...
	go test ./... -count=1
	go test -tags=bdd ./tests/bdd/... -count=1
	cd $(OP_PLUGIN_DIR) && go test ./... -count=1
	mkdir -p $(BIN_DIR)
	go build -v -o $(BIN_PATH) $(CMD)

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) $(CMD)

run:
	go run $(CMD) $(ARGS)

install:
	go install $(CMD)

clean:
	rm -rf $(OUT_DIR) $(BIN_DIR)

op-plugin-test:
	cd $(OP_PLUGIN_DIR) && go test ./... -count=1

op-plugin-build:
	mkdir -p $(BIN_DIR)
	cd $(OP_PLUGIN_DIR) && go build -o ../../$(OP_PLUGIN_BIN) .

op-plugin-install-local: op-plugin-build
	mkdir -p ~/.op/plugins/local
	chmod 700 ~/.op ~/.op/plugins ~/.op/plugins/local
	cp $(OP_PLUGIN_BIN) ~/.op/plugins/local/$(OP_PLUGIN_NAME)
	chmod 755 ~/.op/plugins/local/$(OP_PLUGIN_NAME)

cross-build: clean
	mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD)
	GOOS=darwin GOARCH=arm64 go build -o $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD)
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD)
	GOOS=linux GOARCH=arm64 go build -o $(OUT_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD)
