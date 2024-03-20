# Installation


## Install Go

[Go 1.22+](https://golang.org/doc/install) is required.

## Install the `panacead`

If you want to install the `panacead` binary to run your node or to communicate with other nodes,
please clone the `panacea-core` project and build it.

```bash
# Make sure to checkout the correct branch.
git clone -b v2.2.0 https://github.com/medibloc/panacea-core
cd panacea-core
make install  # All binaries are installed in $GOPATH/bin
```

Verify that the `panacead` binary is installed successfully.
```bash
$ panacead version
2.2.0
```

## Import `panacea-core` as a Go dependency

If you want to develop Go applications by importing the `panacea-core`,
you cannot run `go get github.com/medibloc/panacea-core/v2` directly due to [the design of Go Modules](https://github.com/golang/go/issues/30354)
, which doesn't honor `replace` directives in the `go.mod` of the `panacea-core`.

As a workaround, please add `replace` directives in your `go.mod` as below.
```
module your.com/yours

go 1.22

replace (
    github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
    github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
    
    // If you are using a ledger, you may need to replace the line as shown below:
    github.com/cosmos/ledger-cosmos-go => github.com/cosmos/ledger-cosmos-go v0.12.4
)
```

Then, you can `go get` the `panacea-core`.
```bash
go get github.com/medibloc/panacea-core/v2
```
