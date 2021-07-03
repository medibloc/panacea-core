FROM golang:1.16.5-buster AS build-env

ENV PACKAGES make git gcc wget
RUN apt-get update -y
RUN apt-get install -y $PACKAGES

WORKDIR /src

COPY . .

RUN make clean && make build


FROM debian:buster-slim

COPY --from=build-env /src/build/panacead /usr/bin/panacead
COPY --from=build-env /src/docker/lib/libwasmvm.so /usr/lib/libwasmvm.so

RUN chmod +x /usr/bin/panacead

EXPOSE 26656 26657 1317 9090

CMD ["panacead", "version"]

