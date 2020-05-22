module github.com/kava-labs/kava

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/cosmos/cosmos-sdk v0.38.3
	github.com/gorilla/mux v1.7.3
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.3
	github.com/tendermint/tm-db v0.5.0
	golang.org/x/net v0.0.0-20190930134127-c5a3c61f89f3 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

// patch bech32 decoding
replace github.com/btcsuite/btcutil => github.com/kava-labs/btcutil v0.0.0-20200522184203-886d33430f06
