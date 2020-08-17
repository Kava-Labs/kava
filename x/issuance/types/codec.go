package types

import "github.com/cosmos/cosmos-sdk/codec"

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers the necessary types for issuance module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgIssueTokens{}, "issuance/MsgIssueTokens", nil)
	cdc.RegisterConcrete(MsgRedeemTokens{}, "issuance/MsgRedeemTokens", nil)
	cdc.RegisterConcrete(MsgBlockAddress{}, "issuance/MsgBlockAddress", nil)
	cdc.RegisterConcrete(MsgUnblockAddress{}, "issuance/MsgUnblockAddress", nil)
	cdc.RegisterConcrete(MsgSetPauseStatus{}, "issuance/MsgChangePauseStatus", nil)
	cdc.RegisterConcrete(Asset{}, "issuance/Asset", nil)
}
