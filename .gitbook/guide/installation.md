# Installation

This guide will explain how to install the `panacead` and `panaceacli` entrypoints onto your system.

## Install Go

[Go 1.15+](https://golang.org/doc/install) is required.

## Install the `panacead` and `panaceacli`

Install the latest version of Panacea Core.

```bash
git clone https://github.com/medibloc/panacea-core
cd panacea-core
git checkout v1.3.3
make install  # All binaries are installed in $GOPATH/bin
```

Verify that all binaries are installed successfully.
```bash
$ panacead version --long
name: panacea-core
server_name: panacead
client_name: panaceacli
version: 1.3.3
commit: 2dfceffc647db15499c715e8833c5e379c04e028
build_tags: ' ledger'
go: go version go1.15.5 darwin/amd64

$ panaceacli version --long
name: panacea-core
server_name: panacead
client_name: panaceacli
version: 1.3.3
commit: 2dfceffc647db15499c715e8833c5e379c04e028
build_tags: ' ledger'
go: go version go1.15.5 darwin/amd64
```
