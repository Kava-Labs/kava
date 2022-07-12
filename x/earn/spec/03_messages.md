# Messages

## Ethereum ERC20 to Kava Transfers

Ethereum ERC20 tokens are transferred with the `MsgBridgeEthereumToKava` message
type. Only addresses assigned as a permissioned relayer set in [params] may
submit this message, otherwise the transaction will be rejected.

```go
// MsgBridgeEthereumToKava defines a ERC20 bridge transfer from Ethereum to Kava.
type MsgBridgeEthereumToKava struct {
	Relayer string `protobuf:"bytes,1,opt,name=relayer,proto3" json:"relayer,omitempty"`
	// Originating Ethereum ERC20 contract address
	EthereumERC20Address string `protobuf:"bytes,2,opt,name=ethereum_erc20_address,json=ethereumErc20Address,proto3" json:"ethereum_erc20_address,omitempty"`
	// ERC20 token amount to transfer
	Amount github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,3,opt,name=amount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"amount"`
	// Receiver hex address on Kava
	Receiver string `protobuf:"bytes,4,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// Unique sequence per bridge event
	Sequence github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,5,opt,name=sequence,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"sequence"`
}
```

[params]: 05_params.md
