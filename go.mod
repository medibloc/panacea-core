module github.com/medibloc/panacea-core

go 1.12

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

require (
	github.com/cosmos/cosmos-sdk v0.34.7
	github.com/golangci/golangci-lint v1.19.1 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.9
)

replace github.com/cosmos/cosmos-sdk => github.com/medibloc/cosmos-sdk v0.35.3
