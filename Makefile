VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
GOBIN ?= $(GOPATH)/bin
GOSUM := $(shell which gosum)

export GO111MODULE = on


ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))
build_tags += ledger

# process linker flags

ldflags = -X github.com/medibloc/panacea-core/version.Version=$(VERSION) \
	-X github.com/medibloc/panacea-core/version.Commit=$(COMMIT) \
  -X "github.com/medibloc/panacea-core/version.BuildTags=$(build_tags)"

ifneq ($(GOSUM),)
ldflags += -X github.com/medibloc/panacea-core/version.GoSumHash=$(shell $(GOSUM) go.sum)
endif

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))
BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

all: get_tools install

########################################
### Testing

lint:
	golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify


########################################
### Build/Install

build: go.sum
	go build -mod=readonly $(BUILD_FLAGS) -o build/panacead ./cmd/panacead
	go build -mod=readonly $(BUILD_FLAGS) -o build/panaceacli ./cmd/panaceacli
	go build -mod=readonly $(BUILD_FLAGS) -o build/panaceakeyutil ./cmd/panaceakeyutil

update_panacea_lite_docs:
	@statik -src=client/lcd/swagger-ui -dest=client/lcd -f

install: go.sum update_panacea_lite_docs
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/panacead
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/panaceacli
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/panaceakeyutil

########################################
### Tools & dependencies

get_tools:
	GO111MODULE=off go get github.com/rakyll/statik
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.22.2

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
