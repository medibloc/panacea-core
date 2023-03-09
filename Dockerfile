FROM golang:1.19.2-bullseye AS build-env

# Install minimum necessary dependencies,
ENV PACKAGES make git gcc
RUN apt-get update -y
RUN apt-get install -y $PACKAGES

# Add 'panacea-core' source files
COPY . /src/panacea-core

# Set working directory for the 'panacea-core' build
WORKDIR /src/panacea-core

# Install panacea-core
RUN make clean && make build

# Final image
FROM debian:bullseye-slim

# Copy over binaries from the build-env
COPY --from=build-env /src/panacea-core/build/panacead /usr/bin/panacead

RUN chmod +x /usr/bin/panacead

EXPOSE 26656 26657 1317 9090
