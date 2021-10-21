module github.com/kava-labs/kava

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.44.2
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.2.1
	github.com/tendermint/tendermint v0.34.13
	github.com/tendermint/tm-db v0.6.4
)

replace (
	// patch bech32 decoding to enable larger string lengths // TODO is this needed in sdk v0.44?
	github.com/btcsuite/btcutil => github.com/kava-labs/btcutil v0.0.0-20200522184203-886d33430f06
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
)
