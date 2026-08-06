FROM golang:1.26.5-trixie@sha256:8229e3b2cf7fc08878a86977547e3119c173681c3cc4a64c38cf0c6fe0b42fa8 AS build-env

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
FROM debian:trixie-slim@sha256:3a39a0592364683e6bab97937b72cad5a8fa6dcbbee90edb3bb48c7f8e94f258

# Copy over binaries from the build-env
COPY --from=build-env /src/panacea-core/build/panacead /usr/bin/panacead

RUN chmod +x /usr/bin/panacead

EXPOSE 26656 26657 1317 9090
