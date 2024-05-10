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
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/precisebank/keeper"
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
