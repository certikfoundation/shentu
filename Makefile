PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
GOBIN ?= $(GOPATH)/bin
STATIK = $(GOBIN)/statik
SHASUM := $(shell which sha256sum)
PKG_LIST := $(shell go list ./...)
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf
verbosity = 2

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(GAIA_BUILD_OPTIONS)))
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=certik \
		  -X github.com/cosmos/cosmos-sdk/version.ServerName=certikd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

build_tags := $(strip $(build_tags))
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

# The below include contains the tools target.
include devtools/Makefile

export GO111MODULE = on

all: install release lint test

install: go.sum
	go install $(BUILD_FLAGS) ./cmd/certikd

update-swagger-docs: statik
	$(GOBIN)/statik -src=client/lcd/swagger-ui -dest=client/lcd -f -m
	@if [ -n "$(git status --porcelain)" ]; then \
        echo "\033[91mSwagger docs are out of sync!!!\033[0m";\
        exit 1;\
    else \
    	echo "\033[92mSwagger docs are in sync\033[0m";\
    fi
	@cp -r client/lcd/swagger-ui docs/swagger

update-cli-docs: install
	certikd --doc docs/cli/certikd
	certikcli --doc docs/cli/certikcli
	@perl -pi -e "s|^#* Auto generated by .*||" docs/cli/**/*.md
	@perl -pi -e "s|$$HOME|~|" docs/cli/**/*.md

release: go.sum
	#GOOS=linux go build $(BUILD_FLAGS) -o build/certikd ./cmd/certikd
	#GOOS=windows go build $(BUILD_FLAGS) -o build/certikd.exe ./cmd/certikd
	GOOS=darwin go build $(BUILD_FLAGS) -o build/certikd-macos ./cmd/certikd

build: go.sum
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/certikd.exe ./cmd/certikd
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/certikd ./cmd/certikd
endif

build-linux:
	mkdir -p ./build
	docker build --tag shentu ./
	docker create --name temp shentu:latest
	docker cp temp:/usr/local/bin/certikd ./build/
	docker rm temp

########## Tools ##########

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify

clean:
	rm -rf snapcraft-local.yaml build/

distclean:
	rm -rf \
    gitian-build-darwin/ \
    gitian-build-linux/ \
    gitian-build-windows/ \
    .gitian-builder-cache/

tidy:
	@gofmt -s -w .
	@go mod tidy

lint: tidy
	@GO111MODULE=on golangci-lint run --config .golangci.yml

########## Testing ##########

test: tidy
	@GO111MODULE=on go test ${PKG_LIST}

coverage.out: tidy
	@GO111MODULE=on go test -short -coverprofile=coverage.out -covermode=atomic ${PKG_LIST}

test-cov: coverage.out
	@GO111MODULE=on go tool cover -func $<

test-cov-html: coverage.out
	@GO111MODULE=on go tool cover -html $<

image: Dockerfile Dockerfile.update
	@docker rmi -f shentu-base -f shentu
	@docker build -t shentu-base -t shentu . -f Dockerfile

image.update: Dockerfile.update
	@docker rmi -f shentu
	@docker build -t shentu . -f Dockerfile.update

include .env

localnet: localnet.down image.update docker-compose.yml ./devtools/localnet/localnet_client_setup.sh
	@$(RM) -r ${LOCALNET_ROOT}
	@docker run --volume $(abspath ${LOCALNET_ROOT}):/root --workdir /root -it shentu certikd testnet --v 4 --output-dir /root --starting-ip-address ${LOCALNET_START_IP} --chain-id certikchain
	@docker-compose up -d
	@docker exec $(shell basename $(CURDIR))_client_1 bash /shentu/devtools/localnet/localnet_client_setup.sh

build-docker-certikdnode: build-linux
	$(MAKE) -C networks/local

localnet.client:
	@docker exec -it $(shell basename $(CURDIR))_client_1 bash

localnet.both: localnet localnet.client

localnet.down:
	@docker-compose down --remove-orphans

.PHONY: all install release release32 fix lint test cov coverage coverage.out image image.update localnet localnet.client localnet.both localnet.down

# include simulations
include sims.mk