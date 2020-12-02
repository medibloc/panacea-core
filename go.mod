module github.com/medibloc/panacea-core

go 1.14

require (
	github.com/btcsuite/btcutil v1.0.2
	github.com/cosmos/cosmos-sdk v0.37.14
	github.com/cosmos/go-bip39 v0.0.0-20180618194314-52158e4697b8
	github.com/gorilla/mux v1.7.0
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/pborman/uuid v1.2.1
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.6.1
	github.com/stretchr/testify v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.13
	github.com/tendermint/tm-db v0.2.0
	golang.org/x/crypto v0.0.0-20200115085410-6d4e4cb37c7d
	golang.org/x/net v0.0.0-20190923162816-aa69164e4478 // indirect
	golang.org/x/sys v0.0.0-20190922100055-0a153f010e69 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/cosmos/cosmos-sdk => github.com/medibloc/cosmos-sdk v0.37.14-internal
