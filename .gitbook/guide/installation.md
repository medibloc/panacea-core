# Installation

This guide will explain how to install the `panacead` entrypoints onto your system.

## Install Go

[Go 1.15+](https://golang.org/doc/install) is required.

## Install the `panacead`

Install the latest version of Panacea Core.

```bash
git clone https://github.com/medibloc/panacea-core
cd panacea-core
git checkout v2.0.0-alpha.1
make install  # All binaries are installed in $GOPATH/bin
```

Verify that all binaries are installed successfully.
```bash
$ panacead version --long
name: panacea-core
server_name: <appd>
version: 2.0.0-alpha.1
commit: d771b7b826e3e396222016ec1d2367e51bc79024
build_tags: ' ledger'
go: go version go1.16.3 darwin/amd64
```
