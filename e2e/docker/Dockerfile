FROM golang:1.26.5-trixie@sha256:8229e3b2cf7fc08878a86977547e3119c173681c3cc4a64c38cf0c6fe0b42fa8 AS build-env

ENV GOTOOLCHAIN=local

WORKDIR /src/panacea-core
COPY . ./
COPY --from=panacea_vendor . ./vendor
COPY --from=panacea_e2e_tools faultproxy ./scripts/e2e/faultproxy

ARG PANACEA_VERSION=worktree
ARG PANACEA_COMMIT=unknown
ARG PANACEA_CMT_VERSION=unknown
ARG PANACEA_TM_VERSION=unknown
RUN --network=none --mount=type=cache,target=/root/.cache/go-build \
    set -eu \
    && grep -Eq -- '-mod=(readonly|\$\(GO_BUILD_MOD\))' Makefile \
    && sed -i 's/-mod=readonly/-mod=$(GO_BUILD_MOD)/' Makefile \
    && grep -Fq -- '-mod=$(GO_BUILD_MOD)' Makefile \
    && sed -i \
        -e 's/^export VERSION :=.*/export VERSION ?= worktree/' \
        -e 's/^export COMMIT :=.*/export COMMIT ?= unknown/' \
        -e 's/^export CMTVERSION :=.*/export CMTVERSION ?= unknown/' \
        -e 's/^export TMVERSION :=.*/export TMVERSION ?= unknown/' \
        Makefile \
    && touch go.sum \
    && make build \
        LEDGER_ENABLED=false \
        GO_BUILD_MOD=vendor \
        VERSION="${PANACEA_VERSION}" \
        COMMIT="${PANACEA_COMMIT}" \
        CMTVERSION="${PANACEA_CMT_VERSION}" \
        TMVERSION="${PANACEA_TM_VERSION}" \
    && cd /src/panacea-core/scripts/e2e/faultproxy \
    && CGO_ENABLED=0 GOWORK=off go build -mod=readonly -trimpath \
        -o /src/panacea-e2e-faultproxy .

FROM debian:trixie-slim@sha256:3a39a0592364683e6bab97937b72cad5a8fa6dcbbee90edb3bb48c7f8e94f258

COPY --from=build-env /src/panacea-core/build/panacead /usr/bin/panacead
COPY --from=build-env /src/panacea-e2e-faultproxy /usr/bin/panacea-e2e-faultproxy
RUN chmod 0755 /usr/bin/panacead /usr/bin/panacea-e2e-faultproxy

EXPOSE 26656 26657 1317 9090
