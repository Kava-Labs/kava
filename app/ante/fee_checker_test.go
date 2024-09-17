package ante_test

import (
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	antetestutil "github.com/cosmos/cosmos-sdk/x/auth/ante/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	kavaante "github.com/kava-labs/kava/app/ante"
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
	txBuilder   client.TxBuilder

	accountKeeper  keeper.AccountKeeper
	bankKeeper     *authtestutil.MockBankKeeper
	feeGrantKeeper *antetestutil.MockFeegrantKeeper
	encCfg         moduletestutil.TestEncodingConfig
}

func SetupTestSuite(t *testing.T, isCheckTx bool) *AnteTestSuite {
	t.Helper()

	suite := &AnteTestSuite{}

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

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

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

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			suite.txBuilder, priv, suite.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return suite.txBuilder.GetTx(), nil
}

// TestDeductFees tests the full AnteHandler flow with the DeductFeeDecorator
// using the custom TxFeeChecker
func TestDeductFees(t *testing.T) {
	s := SetupTestSuite(t, false)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	accs := s.CreateTestAccounts(1)

	// msg and signatures
	msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	require.NoError(t, err)

	dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, nil, nil)
	antehandler := sdk.ChainAnteDecorators(dfd)
	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Return(sdkerrors.ErrInsufficientFunds)

	_, err = antehandler(s.ctx, tx, false)
	require.NotNil(t, err, "Tx did not error when fee payer had insufficient funds")

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Return(nil)

	_, err = antehandler(s.ctx, tx, false)
	require.Nil(t, err, "Tx errored after account has been set with sufficient funds")
}

// TestCheckTxFeeWithAnyValidatorMinGasPrices unit test that tests purely only
// the custom TxFeeChecker function
func TestCheckTxFeeWithAnyValidatorMinGasPrices(t *testing.T) {
	tests := []struct {
		name             string
		giveIsCheckTx    bool
		giveMinGasPrices sdk.DecCoins
		giveFee          legacytx.StdFee
		wantErr          string
		wantFeeCoins     sdk.Coins
		wantPriority     int64
	}{
		{
			"single min, single given",
			false,
			sdk.NewDecCoins(sdk.NewDecCoinFromDec("ukava", sdk.MustNewDecFromStr("0.001"))),
			legacytx.NewStdFee(
				100000,
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000)),
			),
			"",
			sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000)),
			0,
		},
		{
			"multi min, single given",
			false,
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("ukava", sdk.MustNewDecFromStr("0.001")),
				sdk.NewDecCoinFromDec("usdt", sdk.MustNewDecFromStr("0.0003")),
			),
			legacytx.NewStdFee(
				100000,
				sdk.NewCoins(sdk.NewInt64Coin("usdt", 1000)),
			),
			"",
			sdk.NewCoins(sdk.NewInt64Coin("usdt", 1000)),
			0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger()).
				WithIsCheckTx(tc.giveIsCheckTx).
				WithMinGasPrices(tc.giveMinGasPrices)

			// Double check ctx was setup properly
			require.Equal(t, tc.giveIsCheckTx, ctx.IsCheckTx())
			require.Equal(t, tc.giveMinGasPrices, ctx.MinGasPrices())

			tx := legacytx.NewStdTx(nil, tc.giveFee, nil, "")

			feeCoins, priority, err := kavaante.CheckTxFeeWithAnyValidatorMinGasPrices(ctx, tx)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantFeeCoins, feeCoins)
			require.Equal(t, tc.wantPriority, priority)
		})
	}
}
