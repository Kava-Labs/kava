module github.com/kava-labs/kava

go 1.13

require (
	github.com/cosmos/cosmos-sdk v0.34.4-0.20190925161702-9d0bed8f4f4e
	github.com/gorilla/mux v1.7.3
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.5
	github.com/tendermint/tm-db v0.2.0
	gopkg.in/yaml.v2 v2.2.4
)

replace github.com/cosmos/cosmos-sdk => ../../cosmos/cosmos-sdk
