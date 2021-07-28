# Installation


## Install Go

[Go 1.15+](https://golang.org/doc/install) is required.

## Install the `panacead`

If you want to install the `panacead` binary to run your node or to communicate with other nodes,
please clone the `panacea-core` project and build it.

```bash
# Make sure to checkout the correct branch.
git clone -b v2.0.1 https://github.com/medibloc/panacea-core
cd panacea-core
make install  # All binaries are installed in $GOPATH/bin
```

Verify that the `panacead` binary is installed successfully.
```bash
$ panacead version --long
name: panacea-core
server_name: <appd>
version: 2.0.1
commit: b36d1dac432c75a6d865e75767fe227a4ca125ca
build_tags: ' ledger'
go: go version go1.16.3 darwin/amd64
```

## Import `panacea-core` as a Go dependency

If you want to develop Go applications by importing the `panacea-core`,
you cannot run `go get github.com/medibloc/panacea-core/v2` directly due to [the design of Go Modules](https://github.com/golang/go/issues/30354)
, which doesn't honor `replace` directives in the `go.mod` of the `panacea-core`.

As a workaround, please add `replace` directives in your `go.mod` as below.
```
module your.com/yours

go 1.16

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
```

Then, you can `go get` the `panacea-core`.
```bash
go get github.com/medibloc/panacea-core/v2
```
