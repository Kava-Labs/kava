module github.com/kava-labs/kava

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.44.3
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/regen-network/cosmos-proto v0.3.1 // indirect
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.14
	github.com/tendermint/tm-db v0.6.4
	google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)

replace (
	github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.34.14
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)

// See https://github.com/cosmos/cosmos-sdk/pull/10401, https://github.com/cosmos/cosmos-sdk/commit/0592ba6158cd0bf49d894be1cef4faeec59e8320
replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0
