package app

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
)

// d is an alias for sdk.MustNewDecFromStr
var d = sdk.MustNewDecFromStr

type tallyHandlerSuite struct {
	suite.Suite
	app TestApp
	ctx sdk.Context

	staking stakingHelper

	tallier TallyHandler
}

func TestTallyHandlerSuite(t *testing.T) {
	suite.Run(t, new(tallyHandlerSuite))
}

func (suite *tallyHandlerSuite) SetupTest() {
	suite.app = NewTestApp()
	suite.app.InitializeFromGenesisStates()
	genesisTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.app.NewContext(false, tmproto.Header{Height: 1, Time: genesisTime})

	suite.staking = stakingHelper{suite.app.GetStakingKeeper()}
	suite.staking.setBondDenom(suite.ctx, "ukava")

	suite.tallier = NewTallyHandler(
		suite.app.GetGovKeeper(),
		suite.app.GetStakingKeeper(),
		suite.app.GetSavingsKeeper(),
		suite.app.GetEarnKeeper(),
		suite.app.GetLiquidKeeper(),
		suite.app.GetBankKeeper(),
	)
}

func (suite *tallyHandlerSuite) TestVotePower_AllSourcesCounted() {
	user := suite.createAccount(suite.newBondCoin(sdk.NewInt(1e9)))

	validator := suite.delegateToNewBondedValidator(user.GetAddress(), sdk.NewInt(1e9))

	derivatives := suite.mintDerivative(user.GetAddress(), validator.GetOperator(), sdk.NewInt(500e6))

	suite.allowBKavaEarnDeposits()
	suite.earnDeposit(
		user.GetAddress(),
		sdk.NewCoin(derivatives.Denom, sdk.NewInt(250e6)),
	)

	proposal := suite.createProposal()
	suite.voteOnProposal(user.GetAddress(), proposal.ProposalId, govtypes.OptionYes)

	_, _, results := suite.tallier.Tally(suite.ctx, proposal)
	suite.Equal(sdk.NewInt(500e6+250e6+250e6), results.Yes)
	suite.Equal(sdk.ZeroInt(), results.No)
	suite.Equal(sdk.ZeroInt(), results.NoWithVeto)
	suite.Equal(sdk.ZeroInt(), results.Abstain)
}

func (suite *tallyHandlerSuite) TestVotePower_UserOverridesValidator() {
	user := suite.createAccount(suite.newBondCoin(sdk.NewInt(1e9)))

	delegated := sdk.NewInt(1e9)
	validator := suite.delegateToNewBondedValidator(user.GetAddress(), delegated)
	selfDelegated := validator.GetTokens().Sub(delegated)

	derivatives := suite.mintDerivative(user.GetAddress(), validator.GetOperator(), sdk.NewInt(500e6))

	suite.allowBKavaEarnDeposits()
	suite.earnDeposit(
		user.GetAddress(),
		sdk.NewCoin(derivatives.Denom, sdk.NewInt(250e6)),
	)

	proposal := suite.createProposal()

	// Validator votes, inheriting user's stake and bkava.
	suite.voteOnProposal(validator.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionYes)

	// use wrapped context to discard the state changes
	readOnlyCtx, _ := suite.ctx.CacheContext()
	_, _, results := suite.tallier.Tally(readOnlyCtx, proposal)
	userPower := sdk.NewInt(500e6 + 250e6 + 250e6)
	suite.Equal(
		selfDelegated.Add(userPower),
		results.Yes,
	)
	suite.Equal(sdk.ZeroInt(), results.No)
	suite.Equal(sdk.ZeroInt(), results.NoWithVeto)
	suite.Equal(sdk.ZeroInt(), results.Abstain)

	// User votes, taking power away from validator.
	suite.voteOnProposal(user.GetAddress(), proposal.ProposalId, govtypes.OptionNo)

	_, _, results = suite.tallier.Tally(suite.ctx, proposal)
	suite.Equal(selfDelegated, results.Yes)
	suite.Equal(userPower, results.No)
	suite.Equal(sdk.ZeroInt(), results.NoWithVeto)
	suite.Equal(sdk.ZeroInt(), results.Abstain)
}

func (suite *tallyHandlerSuite) TestTallyOutcomes() {
	suite.Run("VotedPowerBelowQuorumFails", func() {
		suite.SetupTest()
		suite.setTallyParams(d("0.4"), d("0.5"), d("0.334"))
		proposal := suite.createProposal()

		v1 := suite.createNewBondedValidator(sdk.NewInt(399_999_999))
		suite.createNewBondedValidator(sdk.NewInt(600_000_001))

		suite.voteOnProposal(v1.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionYes)

		passes, burns, tally := suite.tallier.Tally(suite.ctx, proposal)
		suite.Falsef(passes, "expected proposal to fail, tally: %v", tally)
		suite.Truef(burns, "expected desposit to be burned, tally: %v", tally)
	})
	suite.Run("VetoedFails", func() {
		suite.SetupTest()
		suite.setTallyParams(d("0.4"), d("0.5"), d("0.334"))
		proposal := suite.createProposal()

		v1 := suite.createNewBondedValidator(sdk.NewInt(334_000_001))
		v2 := suite.createNewBondedValidator(sdk.NewInt(665_999_999))

		suite.voteOnProposal(v1.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionNoWithVeto)
		suite.voteOnProposal(v2.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionYes)

		passes, burns, tally := suite.tallier.Tally(suite.ctx, proposal)
		suite.Falsef(passes, "expected proposal to fail, tally: %v", tally)
		suite.Truef(burns, "expected desposit to be burned, tally: %v", tally)
	})
	suite.Run("UnvetoedAndYesAboveThresholdPasses", func() {
		suite.SetupTest()
		suite.setTallyParams(d("0.4"), d("0.5"), d("0.334"))
		proposal := suite.createProposal()

		v1 := suite.createNewBondedValidator(sdk.NewInt(900_000_000))
		v2 := suite.createNewBondedValidator(sdk.NewInt(50_000_001))
		v3 := suite.createNewBondedValidator(sdk.NewInt(49_999_999))

		suite.voteOnProposal(v1.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionAbstain)
		suite.voteOnProposal(v2.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionYes)
		suite.voteOnProposal(v3.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionNo)

		passes, burns, tally := suite.tallier.Tally(suite.ctx, proposal)
		suite.Truef(passes, "expected proposal to pass, tally: %v", tally)
		suite.Falsef(burns, "expected desposit to not burn, tally: %v", tally)
	})
	suite.Run("UnvetoedAndYesBelowThresholdFails", func() {
		suite.SetupTest()
		suite.setTallyParams(d("0.4"), d("0.5"), d("0.334"))
		proposal := suite.createProposal()

		v1 := suite.createNewBondedValidator(sdk.NewInt(900_000_000))
		v2 := suite.createNewBondedValidator(sdk.NewInt(49_999_999))
		v3 := suite.createNewBondedValidator(sdk.NewInt(50_000_001))

		suite.voteOnProposal(v1.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionAbstain)
		suite.voteOnProposal(v2.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionYes)
		suite.voteOnProposal(v3.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionNo)

		passes, burns, tally := suite.tallier.Tally(suite.ctx, proposal)
		suite.Falsef(passes, "expected proposal to pass, tally: %v", tally)
		suite.Falsef(burns, "expected desposit to not burn, tally: %v", tally)
	})
	suite.Run("NotEnoughStakeFails", func() {
		suite.SetupTest()
		suite.setTallyParams(d("0.4"), d("0.5"), d("0.334"))
		proposal := suite.createProposal()

		// no stake

		passes, burns, tally := suite.tallier.Tally(suite.ctx, proposal)
		suite.Falsef(passes, "expected proposal to pass, tally: %v", tally)
		suite.Falsef(burns, "expected desposit to not burn, tally: %v", tally)
	})
	suite.Run("UnvetoedAndAllAbstainedFails", func() {
		suite.SetupTest()
		suite.setTallyParams(d("0.4"), d("0.5"), d("0.334"))
		proposal := suite.createProposal()

		v1 := suite.createNewBondedValidator(sdk.NewInt(1e9))

		suite.voteOnProposal(v1.GetOperator().Bytes(), proposal.ProposalId, govtypes.OptionAbstain)

		passes, burns, tally := suite.tallier.Tally(suite.ctx, proposal)
		suite.Falsef(passes, "expected proposal to pass, tally: %v", tally)
		suite.Falsef(burns, "expected desposit to not burn, tally: %v", tally)
	})

}

func (suite *tallyHandlerSuite) setTallyParams(quorum, threshold, veto sdk.Dec) {
	suite.app.GetGovKeeper().SetTallyParams(suite.ctx, govtypes.TallyParams{
		Quorum:        quorum,
		Threshold:     threshold,
		VetoThreshold: veto,
	})
}

func (suite *tallyHandlerSuite) voteOnProposal(voter sdk.AccAddress, proposalID uint64, option govtypes.VoteOption) {
	gk := suite.app.GetGovKeeper()

	err := gk.AddVote(suite.ctx,
		proposalID,
		voter,
		govtypes.NewNonSplitVoteOption(option),
	)
	suite.Require().NoError(err)
}

func (suite *tallyHandlerSuite) createProposal() govtypes.Proposal {
	gk := suite.app.GetGovKeeper()
	deposit := gk.GetDepositParams(suite.ctx).MinDeposit
	proposer := suite.createAccount(deposit...)

	msg, err := govtypes.NewMsgSubmitProposal(
		govtypes.NewTextProposal("a title", "a description"),
		deposit,
		proposer.GetAddress(),
	)
	suite.Require().NoError(err)

	msgServer := govkeeper.NewMsgServerImpl(gk)
	res, err := msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().NoError(err)

	proposal, found := gk.GetProposal(suite.ctx, res.ProposalId)
	if !found {
		panic("proposal not found")
	}
	return proposal
}

func (suite *tallyHandlerSuite) newBondCoin(amount sdk.Int) sdk.Coin {
	return suite.staking.newBondCoin(suite.ctx, amount)
}

func (suite *tallyHandlerSuite) allowBKavaEarnDeposits() {
	ek := suite.app.GetEarnKeeper()
	earnParams := ek.GetParams(suite.ctx)

	vault := earntypes.NewAllowedVault(
		liquidtypes.DefaultDerivativeDenom,
		earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS},
		false,
		nil,
	)

	earnParams.AllowedVaults = append(earnParams.AllowedVaults, vault)
	ek.SetParams(suite.ctx, earnParams)

	sk := suite.app.GetSavingsKeeper()
	savingsParams := sk.GetParams(suite.ctx)
	savingsParams.SupportedDenoms = append(savingsParams.SupportedDenoms, liquidtypes.DefaultDerivativeDenom)
	sk.SetParams(suite.ctx, savingsParams)
}

func (suite *tallyHandlerSuite) earnDeposit(owner sdk.AccAddress, derivative sdk.Coin) {
	ek := suite.app.GetEarnKeeper()

	err := ek.Deposit(suite.ctx, owner, derivative, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)
}

func (suite *tallyHandlerSuite) mintDerivative(owner sdk.AccAddress, validator sdk.ValAddress, amount sdk.Int) sdk.Coin {
	lk := suite.app.GetLiquidKeeper()

	minted, err := lk.MintDerivative(suite.ctx, owner, validator, suite.newBondCoin(amount))
	suite.Require().NoError(err)

	return minted
}

func (suite *tallyHandlerSuite) delegateToNewBondedValidator(delegator sdk.AccAddress, amount sdk.Int) stakingtypes.ValidatorI {
	valAcc := suite.createAccount(suite.newBondCoin(sdk.NewInt(1e9)))
	validator, err := suite.staking.createUnbondedValidator(suite.ctx, valAcc.GetAddress().Bytes(), sdk.NewInt(1e9))
	suite.Require().NoError(err)

	_, err = suite.staking.delegate(suite.ctx, delegator, validator.GetOperator(), amount)
	suite.Require().NoError(err)

	// bond the validator
	sk := suite.app.GetStakingKeeper()
	staking.EndBlocker(suite.ctx, sk)

	validator, found := sk.GetValidator(suite.ctx, validator.GetOperator())
	if !found {
		panic("validator not found")
	}
	return validator
}

func (suite *tallyHandlerSuite) createNewBondedValidator(selfDelegation sdk.Int) stakingtypes.ValidatorI {
	valAcc := suite.createAccount(suite.newBondCoin(selfDelegation))
	validator, err := suite.staking.createUnbondedValidator(suite.ctx, valAcc.GetAddress().Bytes(), selfDelegation)
	suite.Require().NoError(err)

	// bond the validator
	sk := suite.app.GetStakingKeeper()
	staking.EndBlocker(suite.ctx, sk)

	validator, found := sk.GetValidator(suite.ctx, validator.GetOperator())
	if !found {
		panic("validator not found")
	}
	return validator
}

func (suite *tallyHandlerSuite) createAccount(initialBalance ...sdk.Coin) authtypes.AccountI {
	ak := suite.app.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.ctx, RandomAddress())
	ak.SetAccount(suite.ctx, acc)

	err := suite.app.FundAccount(suite.ctx, acc.GetAddress(), initialBalance)
	suite.Require().NoError(err)

	return acc
}

// stakingHelper wraps the staking keeper with helper functions for testing.
type stakingHelper struct {
	keeper stakingkeeper.Keeper
}

func (h stakingHelper) createUnbondedValidator(ctx sdk.Context, address sdk.ValAddress, selfDelegation sdk.Int) (stakingtypes.ValidatorI, error) {
	msg, err := stakingtypes.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		h.newBondCoin(ctx, selfDelegation),
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1e6),
	)
	if err != nil {
		return nil, err
	}

	msgServer := stakingkeeper.NewMsgServerImpl(h.keeper)
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	validator, found := h.keeper.GetValidator(ctx, address)
	if !found {
		panic("validator not found")
	}
	return validator, nil
}

func (h stakingHelper) delegate(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress, amount sdk.Int) (sdk.Dec, error) {
	msg := stakingtypes.NewMsgDelegate(
		delegator,
		validator,
		h.newBondCoin(ctx, amount),
	)

	msgServer := stakingkeeper.NewMsgServerImpl(h.keeper)
	_, err := msgServer.Delegate(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return sdk.Dec{}, err
	}

	del, found := h.keeper.GetDelegation(ctx, delegator, validator)
	if !found {
		panic("delegation not found")
	}
	return del.Shares, nil
}

func (h stakingHelper) newBondCoin(ctx sdk.Context, amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(h.keeper.BondDenom(ctx), amount)
}

func (h stakingHelper) setBondDenom(ctx sdk.Context, denom string) {
	params := h.keeper.GetParams(ctx)
	params.BondDenom = denom
	h.keeper.SetParams(ctx, params)
}
