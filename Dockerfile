FROM golang:1.23.12-bullseye@sha256:161b8513c09cbfa4c174fd32e46eddc5eddf487a43958b9cf8b07d628e9e0f85 AS build-env

ARG PANACEA_VERSION=dev
ARG PANACEA_COMMIT=unknown

# Install minimum necessary dependencies,
ENV PACKAGES make git gcc
ENV GOTOOLCHAIN=local
RUN apt-get update -y
RUN apt-get install -y $PACKAGES

# Add 'panacea-core' source files
COPY . /src/panacea-core

# Set working directory for the 'panacea-core' build
WORKDIR /src/panacea-core

# Install panacea-core. The Docker build context intentionally excludes .git,
# so release metadata must be supplied explicitly instead of being inferred
# from repository files that are not present in the image build.
RUN make VERSION="$PANACEA_VERSION" COMMIT="$PANACEA_COMMIT" clean && \
    make VERSION="$PANACEA_VERSION" COMMIT="$PANACEA_COMMIT" build

# Final image
FROM debian:bullseye-slim@sha256:cba95a21c96c1f5fc2470081829363eed57706634f7dc26e8c6712934303d57a

# Copy over binaries from the build-env
COPY --from=build-env /src/panacea-core/build/panacead /usr/bin/panacead

RUN chmod +x /usr/bin/panacead

EXPOSE 26656 26657 1317 9090
