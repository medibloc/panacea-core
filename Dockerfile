FROM golang:1.16.5-buster AS build-env

# Install minimum necessary dependencies,
ENV PACKAGES make git gcc
RUN apt-get update -y
RUN apt-get install -y $PACKAGES

# Create directory
RUN mkdir -p /src/panacea-core /src/wasmvm

# Add 'panacea-core' source files
COPY . /src/panacea-core

# Set working directory for the 'panacea-core' build
WORKDIR /src/panacea-core

# Install panacea-core
RUN make clean && make build

# Get 'libwasmvm.so' from wasmvm
RUN git clone -b v0.14.0 https://github.com/CosmWasm/wasmvm.git /src/wasmvm

# Final image
FROM debian:buster-slim

# Copy over binaries from the build-env
COPY --from=build-env /src/panacea-core/build/panacead /usr/bin/panacead
# Copy 'libwasmvm.so' liberary from the build-env
COPY --from=build-env /src/wasmvm/api/libwasmvm.so /usr/lib/libwasmvm.so

RUN chmod +x /usr/bin/panacead

EXPOSE 26656 26657 1317 9090