.DEFAULT_GOAL := help
.PHONY: build build-gui build-debug build-tests test test-all test-go vet check build-parser clean help

CXEXE := cx
ifeq ($(OS),Windows_NT)
CXEXE := cx.exe
endif

# Default build includes cipher, cxos and regexp. Opt out with nocipher,
# nocxos, noregexp. ptr64 switches the VM to 64 bit pointers.
TAGS ?=
GUI_TAGS := cxfx

build: ## Build cx into ./bin
	go build -tags="$(TAGS)" -o ./bin/$(CXEXE) ./cmd/cx

build-gui: ## Build cx with the cxfx GL/GLFW/OpenAL package (needs cgo and X11/GL headers)
	go build -tags="$(TAGS) $(GUI_TAGS)" -o ./bin/$(CXEXE) ./cmd/cx

build-debug: ## Build cx without optimizations
	go build -gcflags="all=-N -l" -tags="$(TAGS)" -o ./bin/$(CXEXE) ./cmd/cx

build-tests: build ## Build the cx program test runner
	go build -o ./bin/cxtest ./cmd/cxtest

test: build-tests ## Run the stable CX program tests
	./bin/cxtest --cxpath=./bin/$(CXEXE) --wdir=./tests --log=fail,stderr --disable-tests=gui,issue

test-all: build-tests ## Run all non-GUI CX program tests, including known issues
	./bin/cxtest --cxpath=./bin/$(CXEXE) --wdir=./tests --log=fail,stderr --disable-tests=gui

test-go: ## Run the Go unit tests
	go test ./...

vet: ## go vet with default, ptr64 and cxfx tags
	go vet ./...
	go vet -tags ptr64 ./n001_types ./n006_ast
	go vet -tags cxfx ./n017_cxfx ./n018_init ./cmd/cx

check: vet test-go test ## vet, Go tests and CX program tests

build-parser: ## Regenerate the CX parsers from the .y grammars
	go tool goyacc -o n019_cxpartialparsing/cxpartialparsing.go -v /dev/null n019_cxpartialparsing/cxpartialparsing.y
	go tool goyacc -o n020_cxparsingcompletor/parsingcompletor.go -v /dev/null n020_cxparsingcompletor/parsingcompletor.y
	gofmt -w n019_cxpartialparsing n020_cxparsingcompletor

clean: ## Remove binaries
	rm -rf ./bin

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-14s\033[0m %s\n", $$1, $$2}'
