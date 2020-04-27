package committee_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/types"
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

// NewCommitteeGenesisState marshals a committee genesis state into json for use in initializing test apps.
func NewCommitteeGenesisState(cdc *codec.Codec, gs types.GenesisState) app.GenesisState {
	return app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(gs)}
}

var _ types.PubProposal = UnregisteredPubProposal{}

// UnregisteredPubProposal is a pubproposal type that is not registered on the amino codec.
type UnregisteredPubProposal struct{}

func (UnregisteredPubProposal) GetTitle() string         { return "unregistered" }
func (UnregisteredPubProposal) GetDescription() string   { return "unregistered" }
func (UnregisteredPubProposal) ProposalRoute() string    { return "unregistered" }
func (UnregisteredPubProposal) ProposalType() string     { return "unregistered" }
func (UnregisteredPubProposal) ValidateBasic() sdk.Error { return nil }
func (UnregisteredPubProposal) String() string           { return "unregistered" }
