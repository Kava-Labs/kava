package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func (suite *TypesTestSuite) TestCommitteeChangeProposalMarshals() {

	ccp := CommitteeChangeProposal{
		Title:       "A Title",
		Description: "A description for this committee.",
		NewCommittee: Committee{
			ID:               12,
			Description:      "This committee is for testing.",
			Members:          nil,
			Permissions:      []Permission{ParamChangePermission{}},
			VoteThreshold:    d("0.667"),
			ProposalDuration: time.Hour * 24 * 7,
		},
	}

	appCdc := codec.New()
	// register sdk types in case their needed
	sdk.RegisterCodec(appCdc)
	codec.RegisterCrypto(appCdc)
	codec.RegisterEvidences(appCdc)
	// register committee types
	RegisterCodec(appCdc)

	var ppModuleCdc PubProposal
	suite.NotPanics(func() {
		ModuleCdc.MustUnmarshalBinaryBare(
			ModuleCdc.MustMarshalBinaryBare(PubProposal(ccp)),
			&ppModuleCdc,
		)
	})

	var ppAppCdc PubProposal
	suite.NotPanics(func() {
		appCdc.MustUnmarshalBinaryBare(
			appCdc.MustMarshalBinaryBare(PubProposal(ccp)),
			&ppAppCdc,
		)
	})

	var ppGovCdc govtypes.Content
	suite.NotPanics(func() {
		govtypes.ModuleCdc.MustUnmarshalBinaryBare(
			govtypes.ModuleCdc.MustMarshalBinaryBare(govtypes.Content(ccp)),
			&ppGovCdc,
		)
	})
}
