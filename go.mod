module github.com/medibloc/panacea-core

go 1.13

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

require (
	github.com/bombsimon/wsl v1.2.1 // indirect
	github.com/btcsuite/btcutil v1.0.2
	github.com/cosmos/cosmos-sdk v0.34.7
	github.com/golangci/golangci-lint v1.22.2 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.6.1
	github.com/stretchr/testify v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.9
	golang.org/x/sys v0.0.0-20190911201528-7ad0cfa0b7b5 // indirect
)

replace github.com/cosmos/cosmos-sdk => github.com/medibloc/cosmos-sdk v0.35.3
