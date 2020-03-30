package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterModuleCodec(cdc)
	ModuleCdc = cdc.Seal()
}

// TODO decide if not using gov's Content type would be better

func RegisterModuleCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*gov.Content)(nil), nil) // registering the Content interface on the ModuleCdc will not conflict with gov.
	// Ideally dist and params would register their proposals on here at their init. However don't want to fork them so:
	cdc.RegisterConcrete(distribution.CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal", nil)
	cdc.RegisterConcrete(params.ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal", nil)
	cdc.RegisterConcrete(gov.TextProposal{}, "cosmos-sdk/TextProposal", nil)
	cdc.RegisterConcrete(gov.SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal", nil)

	RegisterAppCodec(cdc)
}

// RegisterCodec registers the necessary types for the module
func RegisterAppCodec(cdc *codec.Codec) {
	// Proposals
	// The app codec needs the gov.Content type registered. This is done by the gov module.
	// Ideally it would registered here as well in case these modules are ever used separately.
	// However amino panics if you register the same interface a second time. So leaving it out for now.
	// cdc.RegisterInterface((*gov.Content)(nil), nil)
	cdc.RegisterConcrete(CommitteeChangeProposal{}, "kava/CommitteeChangeProposal", nil)
	cdc.RegisterConcrete(CommitteeDeleteProposal{}, "kava/CommitteeDeleteProposal", nil)

	// Permissions
	cdc.RegisterInterface((*Permission)(nil), nil)
	cdc.RegisterConcrete(GodPermission{}, "kava/GodPermission", nil)
	cdc.RegisterConcrete(ParamChangePermission{}, "kava/ParamChangePermission", nil)
	cdc.RegisterConcrete(ShutdownPermission{}, "kava/ShutdownPermission", nil)

	// Msgs
	cdc.RegisterConcrete(MsgSubmitProposal{}, "kava/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgVote{}, "kava/MsgVote", nil)
}

// RegisterProposalTypeCodec registers an external proposal content type defined
// in another module for the internal ModuleCdc. This allows the MsgSubmitProposal
// to be correctly Amino encoded and decoded.
func RegisterProposalTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}
