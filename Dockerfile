# For the instruction, please see scripts/run_two_nodes_docker.sh

FROM golang:alpine AS build-env

# Install minimum necessary dependencies,
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3
RUN apk add --no-cache $PACKAGES

# Install minimum Go tools
RUN go get github.com/golangci/golangci-lint/cmd/golangci-lint

# Set working directory for the build
WORKDIR /src

# Add source files
COPY . .

# build panacea-core
RUN make clean && make build


# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /src/build/panacead /usr/bin/panacead
COPY --from=build-env /src/build/panaceacli /usr/bin/panaceacli
COPY --from=build-env /src/build/panaceakeyutil /usr/bin/panaceakeyutil

RUN chmod +x /usr/bin/panacead
RUN chmod +x /usr/bin/panaceacli
RUN chmod +x /usr/bin/panaceakeyutil

EXPOSE 26656 26657 1317 9090

# Run panacead by default
CMD ["panacead"]
