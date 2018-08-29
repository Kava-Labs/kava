package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	//"github.com/stretchr/testify/require"
	"github.com/cosmos/cosmos-sdk/x/bank"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Setup an example app with an in memory DB and the required keepers
// Also create two accounts with 1000KVA
// Could do with refactoring
func createMockApp() (sdk.Context, bank.Keeper, Keeper, []sdk.AccAddress, sdk.Coins) {
	mApp := mock.NewApp() // creates a half complete app
	coinKeeper := bank.NewKeeper(mApp.AccountMapper)

	// create channel keeper
	keyChannel := sdk.NewKVStoreKey("channel")
	channelKeeper := NewKeeper(mApp.Cdc, keyChannel, coinKeeper)
	// add router?
	//mapp.Router().AddRoute("channel", NewHandler(channelKeeper))

	mApp.CompleteSetup([]*sdk.KVStoreKey{keyChannel}) // needs to be called I think to finish setup

	// create some accounts
	numGenAccs := 2 // create two initial accounts
	genAccFunding := sdk.Coins{sdk.NewCoin("KVA", 1000)}
	genAccs, addrs, _, _ := mock.CreateGenAccounts(numGenAccs, genAccFunding)

	// initialize the app with these accounts
	mock.SetGenesis(mApp, genAccs)

	mApp.BeginBlock(abci.RequestBeginBlock{}) // going off other module tests
	ctx := mApp.BaseApp.NewContext(false, abci.Header{})

	return ctx, coinKeeper, channelKeeper, addrs, genAccFunding
}
