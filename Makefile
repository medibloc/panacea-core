export VERSION := $(shell echo $(shell git describe --tags 2>/dev/null) | sed 's/^v//')
export CMTVERSION := $(shell awk '$$1 == "github.com/cometbft/cometbft" { if ($$2 == "=>") version = $$4; else version = $$2 } END { print version }' go.mod)
export COMMIT := $(shell git log -1 --format='%H' 2>/dev/null)
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
HTTPS_GIT := https://github.com/medibloc/panacea-core.git
DOCKER := $(shell which docker)
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)
DOCS_DOMAIN=docs.gopanacea.org

export GO111MODULE = on

# process build tags
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

ifeq (secp,$(findstring secp,$(COSMOS_BUILD_OPTIONS)))
  build_tags += libsecp256k1_sdk
endif

whitespace :=
whitespace += $(whitespace)
comma := ,

ARTIFACT_DIR := artifacts
GO_BUILD_MOD ?= readonly
RELEASE_BUILD_CONTRACT := panacea-linux-static-v1
override RELEASE_BUILD_MOD := vendor
RELEASE_GOARCH ?=
RELEASE_OUTPUT ?= $(BUILDDIR)/panacead

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  $(error cleveldb is not supported in Cosmos SDK v0.50; use goleveldb)
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  $(error badgerdb is not supported in Cosmos SDK v0.50; use goleveldb)
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  build_tags += rocksdb
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  $(error boltdb is not supported in Cosmos SDK v0.50; use goleveldb)
endif

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=panacea-core \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=panacead \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(CMTVERSION)

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
#ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# Check for debug option
ifeq (debug,$(findstring debug,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif

all: build lint test
###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/
build-linux:
	GOOS=linux GOARCH=$(if $(filter aarch64 arm64,$(shell uname -m)),arm64,amd64) LEDGER_ENABLED=false $(MAKE) build

# release-build is the only supported build contract for distributable Linux
# validator binaries and current-version release images. The contract requires
# a materialized vendor directory and callers must disable Ledger while Make
# parses the build-tag configuration.
override release_ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=panacea-core \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=panacead \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=netgo" \
		  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(CMTVERSION) \
		  -w -s
override release_ldflags := $(strip $(release_ldflags))
override RELEASE_BUILD_FLAGS := -tags "netgo" -ldflags '$(release_ldflags)' -trimpath

release-build: go.sum
	@case "$(RELEASE_GOARCH)" in \
		amd64|arm64) ;; \
		*) echo "RELEASE_GOARCH must be amd64 or arm64" >&2; exit 2 ;; \
	esac
	@mkdir -p "$(dir $(RELEASE_OUTPUT))"
	LC_ALL=C TZ=UTC GOENV=off GOTOOLCHAIN=local GOWORK=off \
		GOFLAGS=-buildvcs=false GOEXPERIMENT= GOFIPS140=off GODEBUG= \
		GOOS=linux GOARCH="$(RELEASE_GOARCH)" GOAMD64=v1 GOARM64=v8.0 \
		CGO_ENABLED=0 \
		go build -mod=$(RELEASE_BUILD_MOD) $(RELEASE_BUILD_FLAGS) \
			-o "$(RELEASE_OUTPUT)" ./cmd/panacead

test: proto-gen-test
	mkdir -p $(ARTIFACT_DIR)
	go test -covermode=count -coverprofile=$(ARTIFACT_DIR)/coverage.out ./...
	go tool cover -html=$(ARTIFACT_DIR)/coverage.out -o $(ARTIFACT_DIR)/coverage.html

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	go $@ -mod=$(GO_BUILD_MOD) $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

.PHONY: build build-linux release-build

distclean: clean
clean:
	rm -rf \
    $(BUILDDIR)/ \
    artifacts/ \
    tmp-swagger-gen/

.PHONY: distclean clean


###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_lint_cmd=golangci-lint
golangci_lint_module=github.com/golangci/golangci-lint/v2/cmd/golangci-lint
golangci_version=v2.11.4

lint:
	@echo "--> Running linter"
	@go install $(golangci_lint_module)@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=10m

lint-fix:
	@echo "--> Running linter"
	@go install $(golangci_lint_module)@$(golangci_version)
	@$(golangci_lint_cmd) run --fix --issues-exit-code=0

.PHONY: lint lint-fix

format:
	@go install mvdan.cc/gofumpt@latest
	@go install $(golangci_lint_module)@$(golangci_version)
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name "*.pb.go" -not -name "*.pb.gw.go" -not -name "*.pulsar.go" -not -path "./crypto/keys/secp256k1/*" | xargs gofumpt -w -l
	golangci-lint run --fix
.PHONY: format


###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm --user $$(id -u):$$(id -g) -e HOME=/tmp -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-check-generated:
	@$(MAKE) --no-print-directory proto-lint
	@$(MAKE) --no-print-directory proto-gen-test
	@$(MAKE) --no-print-directory proto-gen
	@if [ -n "$$(git status --porcelain --untracked-files=all -- \
		':(glob)x/**/*.pb.go' ':(glob)x/**/*.pb.gw.go')" ]; then \
		echo "protobuf generation left uncommitted changes:" >&2; \
		git status --short --untracked-files=all -- \
			':(glob)x/**/*.pb.go' ':(glob)x/**/*.pb.gw.go' >&2; \
		exit 1; \
	fi

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./proto/scripts/protocgen.sh

proto-gen-test:
	@bash ./proto/scripts/protocgen_test.sh

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@$(protoImage) sh ./proto/scripts/protoc-swagger-gen.sh

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --exclude-imports --against $(HTTPS_GIT)#branch=main

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update

.PHONY: proto-all proto-check-generated proto-gen proto-gen-test proto-swagger-gen proto-format proto-lint proto-check-breaking proto-update-deps
