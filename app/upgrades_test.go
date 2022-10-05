package app_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kava-labs/kava/app"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	etherminttypes "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type UpgradeTestSuite struct {
	suite.Suite
	App app.TestApp
	Ctx sdk.Context
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	suite.App = app.NewTestApp()

	cdc := suite.App.AppCodec()

	suite.App = suite.App.InitializeFromGenesisStates(
		app.GenesisState{
			minttypes.ModuleName: cdc.MustMarshalJSON(minttypes.NewGenesisState(
				minttypes.DefaultInitialMinter(),
				// Params reflect mainnet params
				minttypes.Params{
					MintDenom:           "ukava",
					InflationRateChange: sdk.NewDecWithPrec(13, 2),
					InflationMax:        sdk.OneDec(),
					InflationMin:        sdk.OneDec(),
					GoalBonded:          sdk.NewDecWithPrec(67, 2),
					BlocksPerYear:       5256000,
				},
			)),
		},
	)

	suite.Ctx = suite.App.NewContext(false, tmproto.Header{Height: 1})
}

func (suite *UpgradeTestSuite) TestUpdateCosmosMintInflation() {
	mintKeeper := suite.App.GetMintKeeper()
	oldParams := mintKeeper.GetParams(suite.Ctx)
	suite.Equal(sdk.OneDec(), oldParams.InflationMin, "initial InflationMin should be 1")
	suite.Equal(sdk.OneDec(), oldParams.InflationMax, "initial InflationMax should be 1")

	// Run migration
	app.UpdateCosmosMintInflation(suite.Ctx, mintKeeper)

	newParams := mintKeeper.GetParams(suite.Ctx)
	suite.NotEqual(oldParams, newParams, "params should be changed after migration")

	suite.Equal(sdk.MustNewDecFromStr("0.75"), sdk.NewDecWithPrec(75, 2), "sdk.NewDecWithPrec(75, 2) should be 0.75")

	suite.Equal(sdk.MustNewDecFromStr("0.75"), newParams.InflationMin, "InflationMin should changed to 0.75")
	suite.Equal(sdk.MustNewDecFromStr("0.75"), newParams.InflationMax, "InflationMax should changed to 0.75")

	// Other parameters should be unchanged
	suite.Equal(oldParams.MintDenom, newParams.MintDenom)
	suite.Equal(oldParams.InflationRateChange, newParams.InflationRateChange)
	suite.Equal(oldParams.GoalBonded, newParams.GoalBonded)
	suite.Equal(oldParams.BlocksPerYear, newParams.BlocksPerYear)
}

func (suite *UpgradeTestSuite) TestUpdateSavingsParams() {
	savingsKeeper := suite.App.GetSavingsKeeper()
	oldParams := savingsKeeper.GetParams(suite.Ctx)
	suite.Empty(oldParams.SupportedDenoms, "initial SupportedDenoms should be empty")

	// Run migration
	app.UpdateSavingsParams(suite.Ctx, savingsKeeper)

	newParams := savingsKeeper.GetParams(suite.Ctx)
	suite.NotEqual(oldParams, newParams, "params should be changed after migration")

	suite.ElementsMatch(
		[]string{
			"ukava",
			"bkava",
			"erc20/multichain/usdc",
		},
		newParams.SupportedDenoms,
		"SupportedDenoms should be updated to include ukava",
	)
}

func (suite *UpgradeTestSuite) TestConvertEOAsToBaseAccount() {
	ak := suite.App.GetAccountKeeper()

	accCount := 0

	// Add all accounts as EthAccount
	app.IterateEOAAddresses(func(addr string) {
		acc, err := sdk.AccAddressFromBech32(addr)
		suite.NoError(err)

		ethAcc := etherminttypes.EthAccount{
			BaseAccount: authtypes.NewBaseAccount(acc, nil, 0, 0),
			CodeHash:    common.Bytes2Hex(evmtypes.EmptyCodeHash),
		}

		ak.SetAccount(suite.Ctx, &ethAcc)
		accCount++
	})

	// Add a contract EthAccount
	contractAcc := etherminttypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress("contract"), nil, 0, 0),
		CodeHash:    common.Bytes2Hex([]byte("contract code hash")),
	}
	ak.SetAccount(suite.Ctx, &contractAcc)

	// Run migration
	suite.NotPanics(func() {
		app.ConvertEOAsToBaseAccount(suite.Ctx, ak)
	})

	accCountAfter := 0

	// Check that accounts are now BaseAccounts
	app.IterateEOAAddresses(func(addrStr string) {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		suite.Require().NoError(err)

		acc := ak.GetAccount(suite.Ctx, addr)
		suite.Require().NoError(err)

		_, ok := acc.(*authtypes.BaseAccount)
		suite.Require().Truef(ok, "account is not an BaseAccount: %T", acc)
		accCountAfter++
	})

	suite.T().Logf("accounts updated: %d", accCountAfter)

	contractAccAfter := ak.GetAccount(suite.Ctx, sdk.AccAddress("contract"))
	suite.Require().NotNil(contractAccAfter)
	suite.Require().Implements(
		(*etherminttypes.EthAccountI)(nil),
		contractAccAfter,
		"contract account should still be an EthAccount",
	)

	suite.Greater(accCount, 0)
	suite.Greater(accCountAfter, 0)
	suite.Equal(accCount, accCountAfter, "account count should be unchanged")
}

func (suite *UpgradeTestSuite) TestAddKavadistFundAccount() {
	ak := suite.App.GetAccountKeeper()
	maccAddr := ak.GetModuleAddress(kavadisttypes.FundModuleAccount)

	dstk := suite.App.GetDistrKeeper()

	communityCoinsBefore := dstk.GetFeePoolCommunityCoins(suite.Ctx)
	suite.T().Logf("community coins before: %s", communityCoinsBefore)

	acc := ak.NewAccountWithAddress(suite.Ctx, maccAddr)
	ak.SetAccount(suite.Ctx, acc)

	bal := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000000000)))
	suite.App.FundAccount(suite.Ctx, maccAddr, bal)

	// Ensure it is a module account prior to migration
	acc = ak.GetAccount(suite.Ctx, maccAddr)
	_, ok := acc.(authtypes.ModuleAccountI)
	suite.Require().Falsef(ok, "account should not a ModuleAccount: %T", acc)

	suite.Require().IsType(&authtypes.BaseAccount{}, acc)

	app.AddKavadistFundAccount(
		suite.Ctx,
		ak,
		suite.App.GetBankKeeper(),
		dstk,
	)

	acc = ak.GetAccount(suite.Ctx, maccAddr)
	suite.Require().Implements((*authtypes.ModuleAccountI)(nil), acc)

	communityCoinsAfter := dstk.GetFeePoolCommunityCoins(suite.Ctx)
	suite.T().Logf("community coins after: %s", communityCoinsAfter)

	suite.Equal(
		communityCoinsBefore.Add(sdk.NewDecCoinsFromCoins(bal...)...),
		communityCoinsAfter,
	)
}
