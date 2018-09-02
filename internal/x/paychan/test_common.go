package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	//"github.com/stretchr/testify/require"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// Setup an example app with an in memory DB and the required keepers
// Also create two accounts with 1000KVA
// Could do with refactoring
func createMockApp(accountSeeds []string) (sdk.Context, bank.Keeper, Keeper, []sdk.AccAddress, []crypto.PubKey, []crypto.PrivKey, sdk.Coins) {
	mApp := mock.NewApp() // creates a half complete app
	coinKeeper := bank.NewKeeper(mApp.AccountMapper)

	// create channel keeper
	keyChannel := sdk.NewKVStoreKey("channel")
	channelKeeper := NewKeeper(mApp.Cdc, keyChannel, coinKeeper)
	// could add router for msg tests
	//mapp.Router().AddRoute("channel", NewHandler(channelKeeper))

	mApp.CompleteSetup([]*sdk.KVStoreKey{keyChannel})

	// create some accounts
	genAccFunding := sdk.Coins{sdk.NewCoin("KVA", 1000)}
	genAccs, addrs, pubKeys, privKeys := createTestGenAccounts(accountSeeds, genAccFunding)

	// initialize the app with these accounts
	mock.SetGenesis(mApp, genAccs)

	mApp.BeginBlock(abci.RequestBeginBlock{}) // going off other module tests
	ctx := mApp.BaseApp.NewContext(false, abci.Header{})

	return ctx, coinKeeper, channelKeeper, addrs, pubKeys, privKeys, genAccFunding
}

// CreateTestGenAccounts deterministically generates genesis accounts loaded with coins, and returns
// their addresses, pubkeys, and privkeys.
func createTestGenAccounts(accountSeeds []string, genCoins sdk.Coins) (genAccs []auth.Account, addrs []sdk.AccAddress, pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {
	for _, seed := range accountSeeds {
		privKey := ed25519.GenPrivKeyFromSecret([]byte(seed))
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		genAcc := &auth.BaseAccount{
			Address: addr,
			Coins:   genCoins,
		}

		genAccs = append(genAccs, genAcc)
		privKeys = append(privKeys, privKey)
		pubKeys = append(pubKeys, pubKey)
		addrs = append(addrs, addr)
	}
	return
}
