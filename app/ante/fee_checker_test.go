package ante_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	antetestutil "github.com/cosmos/cosmos-sdk/x/auth/ante/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	acc  types.AccountI
	priv cryptotypes.PrivKey
}

type AnteTestSuite struct {
	anteHandler sdk.AnteHandler
	ctx         sdk.Context
	clientCtx   client.Context

	accountKeeper  keeper.AccountKeeper
	bankKeeper     *authtestutil.MockBankKeeper
	feeGrantKeeper *antetestutil.MockFeegrantKeeper
	encCfg         moduletestutil.TestEncodingConfig
}

func SetupTestSuite(t *testing.T, isCheckTx bool) *AnteTestSuite {
	t.Helper()
	suite := &AnteTestSuite{}

	ctrl := gomock.NewController(t)
	suite.bankKeeper = authtestutil.NewMockBankKeeper(ctrl)

	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithIsCheckTx(isCheckTx).WithBlockHeight(1) // app.BaseApp.NewContext(isCheckTx, tmproto.Header{}).WithBlockHeight(1)
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})

	maccPerms := map[string][]string{
		"fee_collector":          nil,
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		"multiPerm":              {"burner", "minter", "staking"},
		"random":                 {"random"},
	}

	suite.accountKeeper = keeper.NewAccountKeeper(
		suite.encCfg.Codec, key, types.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix, types.NewModuleAddress("gov").String(),
	)
	suite.accountKeeper.GetModuleAccount(suite.ctx, types.FeeCollectorName)
	err := suite.accountKeeper.SetParams(suite.ctx, types.DefaultParams())
	require.NoError(t, err)

	suite.clientCtx = client.Context{}.
		WithTxConfig(suite.encCfg.TxConfig)

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   suite.accountKeeper,
			BankKeeper:      suite.bankKeeper,
			FeegrantKeeper:  suite.feeGrantKeeper,
			SignModeHandler: suite.encCfg.TxConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)

	require.NoError(t, err)
	suite.anteHandler = anteHandler

	return suite
}

func (suite *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		acc.SetAccountNumber(uint64(i))
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// TestDeductFees tests the full AnteHandler flow with the DeductFeeDecorator
// using the custom TxFeeChecker
func TestDeductFees(t *testing.T) {
	tests := []struct {
		name             string
		setupMocks       func(s *AnteTestSuite)
		giveMinGasPrices sdk.DecCoins
		giveFee          legacytx.StdFee
		wantErr          string
	}{
		{
			"insufficient funds for fee",
			func(s *AnteTestSuite) {
				s.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(sdkerrors.ErrInsufficientFunds)
			},
			sdk.NewDecCoins(sdk.NewDecCoinFromDec("ukava", sdk.MustNewDecFromStr("0.001"))),
			legacytx.NewStdFee(
				100000,
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000)),
			),
			"insufficient funds: insufficient funds",
		},
		{
			"valid fees",
			func(s *AnteTestSuite) {
				s.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(nil)
			},
			sdk.NewDecCoins(sdk.NewDecCoinFromDec("ukava", sdk.MustNewDecFromStr("0.001"))),
			legacytx.NewStdFee(
				100000,
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000)),
			),
			"",
		},
		{
			"valid fees, multiple min gas prices",
			func(s *AnteTestSuite) {
				s.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(nil)
			},
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("ukava", sdk.MustNewDecFromStr("0.001")),
				sdk.NewDecCoinFromDec("usdt", sdk.MustNewDecFromStr("0.003")),
			),
			legacytx.NewStdFee(
				100000,
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000)),
			),
			"",
		},
		{
			"wrong fees coin, multiple min gas prices",
			func(_ *AnteTestSuite) {},
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("ukava", sdk.MustNewDecFromStr("0.001")),
				sdk.NewDecCoinFromDec("usdt", sdk.MustNewDecFromStr("0.003")),
			),
			legacytx.NewStdFee(
				100000,
				sdk.NewCoins(sdk.NewInt64Coin("cats", 1000)),
			),
			"insufficient fees; got: 1000cats required: 100ukava,300usdt: insufficient fee",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := SetupTestSuite(t, false)

			// keys and addresses
			accs := s.CreateTestAccounts(1)

			// msg and signatures
			msg := testdata.NewTestMsg(accs[0].acc.GetAddress())

			// Setup expectations for mocks
			tc.setupMocks(s)

			// Set the minimum gas prices for test & verify it was set
			s.ctx = s.ctx.
				WithMinGasPrices(tc.giveMinGasPrices).
				WithIsCheckTx(true)
			require.Equal(t, tc.giveMinGasPrices, s.ctx.MinGasPrices())
			require.True(t, s.ctx.IsCheckTx(), "expected IsCheckTx to be true to test minimum gas prices")

			// Create transaction with given fee
			txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(msg))

			txBuilder.SetFeeAmount(tc.giveFee.Amount)
			txBuilder.SetGasLimit(tc.giveFee.Gas)

			// NOTE: Transaction is not signed, as it is not checked for this test.
			tx := txBuilder.GetTx()

			dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, nil, nil)
			antehandler := sdk.ChainAnteDecorators(dfd)

			_, err := antehandler(s.ctx, tx, false)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}
