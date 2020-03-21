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
	cdc.RegisterInterface((*gov.Content)(nil), nil) // registering the Content interface on the ModuleCdc will not conflict with gov.
	// TODO ideally dist and params would register their proposals on here at their init. However can't change them so:
	cdc.RegisterConcrete(distribution.CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal", nil)
	cdc.RegisterConcrete(params.ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal", nil)
	cdc.RegisterConcrete(gov.TextProposal{}, "cosmos-sdk/TextProposal", nil)
	cdc.RegisterConcrete(gov.SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal", nil)

	RegisterCodec(cdc)
	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers the necessary types for the module
func RegisterCodec(cdc *codec.Codec) {

	// The app codec needs the gov.Content type registered. This is done by the gov module.
	// Ideally it would registered here as well in case these modules are ever used separately.
	// However amino panics if you register the same interface a second time. So leaving it out for now.

	//cdc.RegisterInterface((*gov.Content)(nil), nil)

	cdc.RegisterConcrete(CommitteeChangeProposal{}, "kava/CommitteeChangeProposal", nil)
	cdc.RegisterConcrete(CommitteeDeleteProposal{}, "kava/CommitteeDeleteProposal", nil)

	cdc.RegisterInterface((*Permission)(nil), nil)
	cdc.RegisterConcrete(GodPermission{}, "kava/GodPermission", nil)
}
