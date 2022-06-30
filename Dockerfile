FROM golang:1.17-alpine3.14 AS build-env

# Install minimum necessary dependencies,
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git
RUN apk add libusb-dev linux-headers

# Create directory
RUN mkdir -p /src/panacea-core /src/wasmvm

# Add 'panacea-core' source files
COPY . /src/panacea-core

# Set working directory for the 'panacea-core' build
WORKDIR /src/panacea-core

# Get 'libwasmvm.so' from wasmvm
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta7/libwasmvm_muslc.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep d0152067a5609bfdfb3f0d5d6c0f2760f79d5f2cd7fd8513cafa9932d22eb350

RUN make clean && BUILD_TAGS=muslc make build

# Final image
FROM debian:buster-slim
#
## Copy over binaries from the build-env
COPY --from=build-env /src/panacea-core/build/panacead /usr/bin/panacead
#
RUN chmod +x /usr/bin/panacead
#
EXPOSE 26656 26657 1317 9090