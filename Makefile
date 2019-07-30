VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
GOSUM := $(shell which gosum)


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


all: lint install

install: go.sum
		go install -mod=readonly $(BUILD_FLAGS) ./cmd/panacead
		go install -mod=readonly $(BUILD_FLAGS) ./cmd/panaceacli
		go install -mod=readonly $(BUILD_FLAGS) ./cmd/panaceakeyutil

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

lint:
	golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify
