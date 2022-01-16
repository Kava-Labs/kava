module github.com/kava-labs/kava

go 1.17

require (
	github.com/cosmos/cosmos-proto v0.0.0-20211020182451-c7ca7198c2f8
	github.com/cosmos/cosmos-sdk v0.44.5
	github.com/cosmos/ibc-go/v2 v2.0.2
	github.com/gogo/protobuf v1.3.3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.15
	github.com/tendermint/tm-db v0.6.6
	google.golang.org/genproto v0.0.0-20210917145530-b395a37504d4
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// Use the cosmos keyring code
	github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
	// Use patched version based on v0.44.5 - note: not state compatiable
	github.com/cosmos/cosmos-sdk => github.com/kava-labs/cosmos-sdk v0.44.5-kava.1
	// See https://github.com/cosmos/cosmos-sdk/pull/10401, https://github.com/cosmos/cosmos-sdk/commit/0592ba6158cd0bf49d894be1cef4faeec59e8320
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0
	// Use the cosmos modified protobufs
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	// Fix rocksdb for high traffic situations
	github.com/tecbot/gorocksdb => github.com/cosmos/gorocksdb v1.2.0
	// Make sure that only one version of tendermint is imported
	github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.34.15
	// Make sure that we use grpc compatible with cosmos
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
