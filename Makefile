COLOR_OK=\\x1b[0;32m
COLOR_NONE=\x1b[0m
COLOR_ERROR=\x1b[31;01m
COLOR_WARNING=\x1b[33;05m
COLOR_ZSCALER=\x1B[34;01m
GOFMT := gofumpt
GOIMPORTS := goimports

help:
	@echo "$(COLOR_ZSCALER)"
	@echo "  ______              _           "
	@echo " |___  /             | |          "
	@echo "    / / ___  ___ __ _| | ___ _ __ "
	@echo "   / / / __|/ __/ _\` | |/ _ \ '__|"
	@echo "  / /__\__ \ (_| (_| | |  __/ |   "
	@echo " /_____|___/\___\__,_|_|\___|_|   "
	@echo "                                  "
	@echo "                                  "
	@echo "$(COLOR_OK)Zscaler SDK for Golang$(COLOR_NONE)"
	@echo ""
	@echo "$(COLOR_WARNING)Usage:$(COLOR_NONE)"
	@echo "$(COLOR_OK)  make [command]$(COLOR_NONE)"
	@echo ""
	@echo "$(COLOR_WARNING)Available commands:$(COLOR_NONE)"
	@echo "$(COLOR_OK)  build                 Clean and build the Zscaler Golang SDK generated files$(COLOR_NONE)"
	@echo "$(COLOR_WARNING)test$(COLOR_NONE)"
	@echo "$(COLOR_OK)  test:all              Run all tests$(COLOR_NONE)"
	@echo "$(COLOR_OK)  test:zcon        	Run only zcon integration tests$(COLOR_NONE)"
	@echo "$(COLOR_OK)  test:zia        	Run only zpa integration tests$(COLOR_NONE)"
	@echo "$(COLOR_OK)  test:zpa        	Run only zpa integration tests$(COLOR_NONE)"
	@echo "$(COLOR_OK)  test:unit             Run only unit tests$(COLOR_NONE)"

build:
	@echo "$(COLOR_ZSCALER)Building SDK...$(COLOR_NONE)"
	make test:all

test:
	make test:all

test\:all:
	@echo "$(COLOR_ZSCALER)Running all tests...$(COLOR_NONE)"
	@make test:zpa
	@make test:zia

test\:integration\:zcon:
	@echo "$(COLOR_ZSCALER)Running zcon integration tests...$(COLOR_NONE)"
	go test -failfast -race ./zcon/... -race -coverprofile zconcoverage.txt -covermode=atomic -v -parallel 20 -timeout 120m
	go tool cover -func zconcoverage.txt | grep total:
	rm -rf zconcoverage.txt

test\:integration\:zpa:
	@echo "$(COLOR_ZSCALER)Running zpa integration tests...$(COLOR_NONE)"
	go test -failfast -race ./zpa/... -race -coverprofile zpacoverage.txt -covermode=atomic -v -parallel 20 -timeout 120m
	go tool cover -func zpacoverage.txt | grep total:
	rm -rf zpacoverage.txt

test\:integration\:zia:
	@echo "$(COLOR_ZSCALER)Running zia integration tests...$(COLOR_NONE)"
	go test -failfast -race ./zia/... -race -coverprofile ziacoverage.txt -covermode=atomic -v -parallel 10 -timeout 120m
	go tool cover -func ziacoverage.txt | grep total:
	rm -rf zpacoverage.txt

test\:unit:
	@echo "$(COLOR_OK)Running unit tests...$(COLOR_NONE)"
	go test -failfast -race ./tests/unit -test.v

test\:unit\zcon:
	@echo "$(COLOR_OK)Running unit tests...$(COLOR_NONE)"
	go test -failfast -race ./tests/unit/zcon -test.v

test\:unit\:zia:
	@echo "$(COLOR_OK)Running unit tests...$(COLOR_NONE)"
	go test -failfast -race ./tests/unit/zia -test.v

test\:unit\:zpa:
	@echo "$(COLOR_OK)Running unit tests...$(COLOR_NONE)"
	go test -failfast -race ./tests/unit/zpa -test.v

test\:unit\all:
	@echo "$(COLOR_OK)Running unit tests...$(COLOR_NONE)"
	go test -race ./tests/unit/zcon -test.v
	go test -race ./tests/unit/zia -test.v
	go test -race ./tests/unit/zpa -test.v


ziaActivator: GOOS=$(shell go env GOOS)
ziaActivator: GOARCH=$(shell go env GOARCH)
ifeq ($(OS),Windows_NT)  # is Windows_NT on XP, 2000, 7, Vista, 10...
ziaActivator: DESTINATION=C:\Windows\System32
else
ziaActivator: DESTINATION=/usr/local/bin
endif
ziaActivator:
	@echo "==> Installing ziaActivator cli $(DESTINATION)"
	cd ./zia/activation_cli
	go mod vendor && go mod tidy
	@mkdir -p $(DESTINATION)
	@rm -f $(DESTINATION)/ziaActivator
	@go build -o $(DESTINATION)/ziaActivator ./zia/activation_cli/ziaActivator.go
	ziaActivator

zconActivator: GOOS=$(shell go env GOOS)
zconActivator: GOARCH=$(shell go env GOARCH)
ifeq ($(OS),Windows_NT)  # is Windows_NT on XP, 2000, 7, Vista, 10...
zconActivator: DESTINATION=C:\Windows\System32
else
zconActivator: DESTINATION=/usr/local/bin
endif
zconActivator:
	@echo "==> Installing zconActivator cli $(DESTINATION)"
	cd ./zcon/services/activation_cli
	go mod vendor && go mod tidy
	@mkdir -p $(DESTINATION)
	@rm -f $(DESTINATION)/ziaActivator
	@go build -o $(DESTINATION)/zconActivator ./zcon/services/activation_cli/zconActivator.go
	zconActivator

.PHONY: fmt
fmt: check-fmt # Format the code
	@$(GOFMT) -l -w $$(find . -name '*.go' |grep -v vendor) > /dev/null

check-fmt:
	@which $(GOFMT) > /dev/null || GO111MODULE=on go install mvdan.cc/gofumpt@latest

.PHONY: import
import: check-goimports
	@$(GOIMPORTS) -w $$(find . -path ./vendor -prune -o -name '*.go' -print) > /dev/null

check-goimports:
	@which $(GOIMPORTS) > /dev/null || GO111MODULE=on go install golang.org/x/tools/cmd/goimports@latest
