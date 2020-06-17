module github.com/kava-labs/kava

go 1.13

require (
	github.com/99designs/keyring v1.1.5 // indirect
	github.com/cosmos/cosmos-sdk v0.38.4
	github.com/gibson042/canonicaljson-go v1.0.3 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.3 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/handlers v1.4.2 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/otiai10/copy v1.2.0 // indirect
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/rakyll/statik v0.1.7 // indirect
	github.com/regen-network/cosmos-proto v0.3.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.5.1
	github.com/tendermint/go-amino v0.15.1 // indirect
	github.com/tendermint/iavl v0.13.3 // indirect
	github.com/tendermint/tendermint v0.33.5
	github.com/tendermint/tm-db v0.5.1
	gopkg.in/yaml.v2 v2.3.0
)

// patch bech32 decoding to enable larger string lengths
replace github.com/btcsuite/btcutil => github.com/kava-labs/btcutil v0.0.0-20200522184203-886d33430f06
replace github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.34.4-0.20200528144628-f8bad078b7b3