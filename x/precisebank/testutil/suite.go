package testutil

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cometbft/cometbft/version"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/types"
)

type Suite struct {
	suite.Suite

	App           app.TestApp
	Ctx           sdk.Context
	BankKeeper    bankkeeper.Keeper
	AccountKeeper authkeeper.AccountKeeper
	Keeper        keeper.Keeper
}

func (suite *Suite) SetupTest() {
	tApp := app.NewTestApp()

	suite.Ctx = tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.App = tApp
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.AccountKeeper = tApp.GetAccountKeeper()
	suite.Keeper = tApp.GetPrecisebankKeeper()

	cdc := suite.App.AppCodec()
	coins := sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000_000_000_000_000_000))
	authGS := app.NewFundedGenStateWithSameCoins(cdc, coins, []sdk.AccAddress{})

	gs := app.GenesisState{}
	suite.App.InitializeFromGenesisStates(authGS, gs)

	// consensus key - needed to set up evm module
	consPriv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	consAddress := sdk.ConsAddress(consPriv.PubKey().Address())

	// InitializeFromGenesisStates commits first block so we start at 2 here
	suite.Ctx = suite.App.NewContext(false, tmproto.Header{
		Height:          suite.App.LastBlockHeight() + 1,
		ChainID:         app.TestChainId,
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})
}

func (suite *Suite) Commit() {
	_ = suite.App.Commit()
	header := suite.Ctx.BlockHeader()
	header.Height += 1
	suite.App.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.Ctx = suite.App.NewContext(false, header)
}

// MintToAccount mints coins to an account with the x/precisebank methods. This
// must be used when minting extended coins, ie. akava coins. This depends on
// the methods to be properly tested to be implemented correctly.
func (suite *Suite) MintToAccount(addr sdk.AccAddress, amt sdk.Coins) {
	accBalancesBefore := suite.GetAllBalances(addr)

	err := suite.Keeper.MintCoins(suite.Ctx, minttypes.ModuleName, amt)
	suite.Require().NoError(err)

	err = suite.Keeper.SendCoinsFromModuleToAccount(suite.Ctx, minttypes.ModuleName, addr, amt)
	suite.Require().NoError(err)

	// Double check balances are correctly minted and sent to account
	accBalancesAfter := suite.GetAllBalances(addr)

	netIncrease := accBalancesAfter.Sub(accBalancesBefore...)
	suite.Require().Equal(ConvertCoinsToExtendedCoinDenom(amt), netIncrease)

	suite.T().Logf("minted %s to %s", amt, addr)
}

// GetAllBalances returns all the account balances for the given account address.
// This returns the extended coin balance if the account has a non-zero balance,
// WITHOUT the integer coin balance.
func (suite *Suite) GetAllBalances(addr sdk.AccAddress) sdk.Coins {
	// Get all balances for an account
	bankBalances := suite.BankKeeper.GetAllBalances(suite.Ctx, addr)

	// Remove integer coins from the balance
	for _, coin := range bankBalances {
		if coin.Denom == types.IntegerCoinDenom {
			bankBalances = bankBalances.Sub(coin)
		}
	}

	// Replace the integer coin with the extended coin, from x/precisebank
	extendedBal := suite.Keeper.GetBalance(suite.Ctx, addr, types.ExtendedCoinDenom)

	return bankBalances.Add(extendedBal)
}

// ConvertCoinsToExtendedCoinDenom converts sdk.Coins that includes Integer denoms
// to sdk.Coins that includes Extended denoms of the same amount. This is useful
// for testing to make sure only extended amounts are compared instead of double
// counting balances.
func ConvertCoinsToExtendedCoinDenom(coins sdk.Coins) sdk.Coins {
	integerCoinAmt := coins.AmountOf(types.IntegerCoinDenom)
	if integerCoinAmt.IsZero() {
		return coins
	}

	// Remove the integer coin from the coins
	integerCoin := sdk.NewCoin(types.IntegerCoinDenom, integerCoinAmt)

	// Add the equivalent extended coin to the coins
	extendedCoin := sdk.NewCoin(types.ExtendedCoinDenom, integerCoinAmt.Mul(types.ConversionFactor()))

	return coins.Sub(integerCoin).Add(extendedCoin)
}
