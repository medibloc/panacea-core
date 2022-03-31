module github.com/medibloc/panacea-core/v2

go 1.15

require (
	github.com/CosmWasm/wasmd v0.17.0
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/cosmos/cosmos-sdk v0.42.11
	github.com/cosmos/go-bip39 v1.0.0
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/kr/text v0.2.0 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_golang v1.10.0
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.14
	github.com/tendermint/tm-db v0.6.4
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	google.golang.org/genproto v0.0.0-20210207032614-bba0dbe2a9ea
	google.golang.org/grpc v1.42.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

// TODO: use the correct version after releasing medibloc/cosmos-sdk
replace github.com/cosmos/cosmos-sdk => github.com/medibloc/cosmos-sdk v0.42.11-panacea-min-commission-rate-28eaec9
