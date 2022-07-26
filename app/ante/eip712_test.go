package ante_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/ethereum/eip712"
	"github.com/tharsis/ethermint/tests"
	etherminttypes "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	evmutilkeeper "github.com/kava-labs/kava/x/evmutil/keeper"
	evmutiltestutil "github.com/kava-labs/kava/x/evmutil/testutil"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

const (
	ChainID       = "kavatest_1-1"
	USDCCoinDenom = "erc20/usdc"
	USDCCDPType   = "erc20-usdc"
)

type EIP712TestSuite struct {
	suite.Suite

	tApp          app.TestApp
	ctx           sdk.Context
	evmutilKeeper evmutilkeeper.Keeper
	clientCtx     client.Context
	ethSigner     ethtypes.Signer
	testAddr      sdk.AccAddress
	testAddr2     sdk.AccAddress
	testPrivKey   cryptotypes.PrivKey
	testPrivKey2  cryptotypes.PrivKey
	testEVMAddr   evmutiltypes.InternalEVMAddress
	testEVMAddr2  evmutiltypes.InternalEVMAddress
	usdcEVMAddr   evmutiltypes.InternalEVMAddress
}

func (suite *EIP712TestSuite) getEVMAmount(amount int64) sdk.Int {
	incr := sdk.RelativePow(sdk.NewUint(10), sdk.NewUint(18), sdk.OneUint())
	return sdk.NewInt(amount).Mul(sdk.NewIntFromUint64(incr.Uint64()))
}

func (suite *EIP712TestSuite) createTestEIP712CosmosTxBuilder(
	from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins, msgs []sdk.Msg,
) client.TxBuilder {
	var err error

	nonce, err := suite.tApp.GetAccountKeeper().GetSequence(suite.ctx, from)
	suite.Require().NoError(err)

	pc, err := etherminttypes.ParseChainID(chainId)
	suite.Require().NoError(err)
	ethChainId := pc.Uint64()

	// GenerateTypedData TypedData
	fee := legacytx.NewStdFee(gas, gasAmount)
	accNumber := suite.tApp.GetAccountKeeper().GetAccount(suite.ctx, from).GetAccountNumber()

	data := eip712.EIP712SignBytes(chainId, accNumber, nonce, 0, fee, msgs, "")
	typedData, err := eip712.WrapTxToTypedData(ethChainId, msgs, data, &eip712.FeeDelegationOptions{
		FeePayer: from,
	}, suite.tApp.GetEvmKeeper().GetParams(suite.ctx))
	suite.Require().NoError(err)
	sigHash, err := eip712.ComputeTypedDataHash(typedData)
	suite.Require().NoError(err)

	// Sign typedData
	keyringSigner := tests.NewSigner(priv)
	signature, pubKey, err := keyringSigner.SignByAddress(from, sigHash)
	suite.Require().NoError(err)
	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper

	// Add ExtensionOptionsWeb3Tx extension
	var option *codectypes.Any
	option, err = codectypes.NewAnyWithValue(&etherminttypes.ExtensionOptionsWeb3Tx{
		FeePayer:         from.String(),
		TypedDataChainID: ethChainId,
		FeePayerSig:      signature,
	})
	suite.Require().NoError(err)

	suite.clientCtx.TxConfig.SignModeHandler()
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	builder.SetExtensionOptions(option)
	builder.SetFeeAmount(gasAmount)
	builder.SetGasLimit(gas)

	sigsV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		},
		Sequence: nonce,
	}

	err = builder.SetSignatures(sigsV2)
	suite.Require().NoError(err)

	err = builder.SetMsgs(msgs...)
	suite.Require().NoError(err)

	return builder
}

func (suite *EIP712TestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.tApp = tApp
	cdc := tApp.AppCodec()
	suite.evmutilKeeper = tApp.GetEvmutilKeeper()

	addr, privkey := tests.NewAddrKey()
	suite.testAddr = sdk.AccAddress(addr.Bytes())
	suite.testPrivKey = privkey
	suite.testEVMAddr = evmutiltestutil.MustNewInternalEVMAddressFromString(addr.String())
	addr2, privKey2 := tests.NewAddrKey()
	suite.testPrivKey2 = privKey2
	suite.testAddr2 = sdk.AccAddress(addr2.Bytes())
	suite.testEVMAddr2 = evmutiltestutil.MustNewInternalEVMAddressFromString(addr2.String())

	encodingConfig := app.MakeEncodingConfig()
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(tApp.GetEvmKeeper().ChainID())

	// Genesis states
	evmGs := evmtypes.NewGenesisState(
		evmtypes.NewParams("akava", true, true, evmtypes.DefaultChainConfig()),
		nil,
	)

	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	cdpGenState := cdptypes.DefaultGenesisState()
	cdpGenState.Params.GlobalDebtLimit = sdk.NewInt64Coin("usdx", 53000000000000)
	cdpGenState.Params.CollateralParams = cdptypes.CollateralParams{
		{
			Denom:                            USDCCoinDenom,
			Type:                             USDCCDPType,
			LiquidationRatio:                 sdk.MustNewDecFromStr("1.01"),
			DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
			StabilityFee:                     sdk.OneDec(),
			AuctionSize:                      sdk.NewIntFromUint64(10000000000),
			LiquidationPenalty:               sdk.MustNewDecFromStr("0.05"),
			CheckCollateralizationIndexCount: sdk.NewInt(10),
			KeeperRewardPercentage:           sdk.MustNewDecFromStr("0.01"),
			SpotMarketID:                     "usdc:usd",
			LiquidationMarketID:              "usdc:usd:30",
			ConversionFactor:                 sdk.NewInt(18),
		},
	}

	hardGenState := hardtypes.DefaultGenesisState()
	hardGenState.Params.MoneyMarkets = []hardtypes.MoneyMarket{
		{
			Denom: "usdx",
			BorrowLimit: hardtypes.BorrowLimit{
				HasMaxLimit:  true,
				MaximumLimit: sdk.MustNewDecFromStr("100000000000"),
				LoanToValue:  sdk.MustNewDecFromStr("1"),
			},
			SpotMarketID:     "usdx:usd",
			ConversionFactor: sdk.NewInt(1_000_000),
			InterestRateModel: hardtypes.InterestRateModel{
				BaseRateAPY:    sdk.MustNewDecFromStr("0.05"),
				BaseMultiplier: sdk.MustNewDecFromStr("2"),
				Kink:           sdk.MustNewDecFromStr("0.8"),
				JumpMultiplier: sdk.MustNewDecFromStr("10"),
			},
			ReserveFactor:          sdk.MustNewDecFromStr("0.05"),
			KeeperRewardPercentage: sdk.ZeroDec(),
		},
	}

	pricefeedGenState := pricefeedtypes.DefaultGenesisState()
	pricefeedGenState.Params.Markets = []pricefeedtypes.Market{
		{
			MarketID:   "usdx:usd",
			BaseAsset:  "usdx",
			QuoteAsset: "usd",
			Oracles:    []sdk.AccAddress{},
			Active:     true,
		},
		{
			MarketID:   "usdc:usd",
			BaseAsset:  "usdc",
			QuoteAsset: "usd",
			Oracles:    []sdk.AccAddress{},
			Active:     true,
		},
		{
			MarketID:   "usdc:usd:30",
			BaseAsset:  "usdc",
			QuoteAsset: "usd",
			Oracles:    []sdk.AccAddress{},
			Active:     true,
		},
	}
	pricefeedGenState.PostedPrices = []pricefeedtypes.PostedPrice{
		{
			MarketID:      "usdx:usd",
			OracleAddress: sdk.AccAddress{},
			Price:         sdk.MustNewDecFromStr("1.00"),
			Expiry:        time.Now().Add(1 * time.Hour),
		},
		{
			MarketID:      "usdc:usd",
			OracleAddress: sdk.AccAddress{},
			Price:         sdk.MustNewDecFromStr("1.00"),
			Expiry:        time.Now().Add(1 * time.Hour),
		},
		{
			MarketID:      "usdc:usd:30",
			OracleAddress: sdk.AccAddress{},
			Price:         sdk.MustNewDecFromStr("1.00"),
			Expiry:        time.Now().Add(1 * time.Hour),
		},
	}

	genState := app.GenesisState{
		evmtypes.ModuleName:       cdc.MustMarshalJSON(evmGs),
		feemarkettypes.ModuleName: cdc.MustMarshalJSON(feemarketGenesis),
		cdptypes.ModuleName:       cdc.MustMarshalJSON(&cdpGenState),
		hardtypes.ModuleName:      cdc.MustMarshalJSON(&hardGenState),
		pricefeedtypes.ModuleName: cdc.MustMarshalJSON(&pricefeedGenState),
	}

	// funds our test accounts with some ukava
	coinsGenState := app.NewFundedGenStateWithSameCoins(
		tApp.AppCodec(),
		sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e9)),
		[]sdk.AccAddress{suite.testAddr, suite.testAddr2},
	)

	tApp.InitializeFromGenesisStatesWithTimeAndChainID(
		time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		ChainID,
		genState,
		coinsGenState,
	)

	// consensus key
	consPriv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	consAddress := sdk.ConsAddress(consPriv.PubKey().Address())

	ctx := tApp.NewContext(false, tmproto.Header{
		Height:          tApp.LastBlockHeight() + 1,
		ChainID:         ChainID,
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
	suite.ctx = ctx

	// We need to set the validator as calling the EVM looks up the validator address
	// https://github.com/tharsis/ethermint/blob/f21592ebfe74da7590eb42ed926dae970b2a9a3f/x/evm/keeper/state_transition.go#L487
	// evmkeeper.EVMConfig() will return error "failed to load evm config" if not set
	valAcc := &etherminttypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(consAddress.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
	tApp.GetAccountKeeper().SetAccount(ctx, valAcc)
	_, testAddresses := app.GeneratePrivKeyAddressPairs(1)
	valAddr := sdk.ValAddress(testAddresses[0].Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, consPriv.PubKey(), stakingtypes.Description{})
	suite.Require().NoError(err)
	err = tApp.GetStakingKeeper().SetValidatorByConsAddr(ctx, validator)
	suite.Require().NoError(err)
	tApp.GetStakingKeeper().SetValidator(ctx, validator)

	// Deploy an ERC20 contract for USDC
	contractAddr := suite.deployUSDCERC20(tApp, ctx)
	pair := evmutiltypes.NewConversionPair(
		contractAddr,
		USDCCoinDenom,
	)
	suite.usdcEVMAddr = pair.GetAddress()

	// Add a contract to evmutil conversion pair
	suite.evmutilKeeper.SetParams(suite.ctx, evmutiltypes.NewParams(
		evmutiltypes.NewConversionPairs(
			evmutiltypes.NewConversionPair(
				// First contract evmutil module deploys
				evmutiltestutil.MustNewInternalEVMAddressFromString("0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3"),
				"erc20/usdc",
			),
		),
	))

	// allow msgs through evm eip712
	evmKeeper := suite.tApp.GetEvmKeeper()
	params := evmKeeper.GetParams(suite.ctx)
	params.EIP712AllowedMsgs = []evmtypes.EIP712AllowedMsg{
		{
			LegacyMsgType: "evmutil_convert_erc20_to_coin",
			ValueTypes: []evmtypes.EIP712MsgAttrType{
				{Name: "initiator", Type: "string"},
				{Name: "receiver", Type: "string"},
				{Name: "kava_erc20_address", Type: "string"},
				{Name: "amount", Type: "string"},
			},
		},
		{
			LegacyMsgType: "create_cdp",
			ValueTypes: []evmtypes.EIP712MsgAttrType{
				{Name: "sender", Type: "string"},
				{Name: "collateral", Type: "Coin"},
				{Name: "principal", Type: "Coin"},
				{Name: "collateral_type", Type: "string"},
			},
		},
		{
			LegacyMsgType: "hard_deposit",
			ValueTypes: []evmtypes.EIP712MsgAttrType{
				{Name: "depositor", Type: "string"},
				{Name: "amount", Type: "Coin[]"},
			},
		},
	}
	evmKeeper.SetParams(suite.ctx, params)

	// give test address 50k erc20 usdc to begin with
	initBal := suite.getEVMAmount(50_000)
	err = suite.evmutilKeeper.MintERC20(
		ctx,
		pair.GetAddress(), // contractAddr
		suite.testEVMAddr, //receiver
		initBal.BigInt(),
	)
	suite.Require().NoError(err)
	err = suite.evmutilKeeper.MintERC20(
		ctx,
		pair.GetAddress(),  // contractAddr
		suite.testEVMAddr2, //receiver
		initBal.BigInt(),
	)
	suite.Require().NoError(err)

	// We need to commit so that the ethermint feemarket beginblock runs to set the minfee
	// feeMarketKeeper.GetBaseFee() will return nil otherwise
	suite.Commit()

	// set base fee
	suite.tApp.GetFeeMarketKeeper().SetBaseFee(suite.ctx, big.NewInt(100))
}

func (suite *EIP712TestSuite) Commit() {
	_ = suite.tApp.Commit()
	header := suite.ctx.BlockHeader()
	header.Height += 1
	suite.tApp.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.tApp.NewContext(false, header)
}

func (suite *EIP712TestSuite) deployUSDCERC20(app app.TestApp, ctx sdk.Context) evmutiltypes.InternalEVMAddress {
	// make sure module account is created
	suite.tApp.FundModuleAccount(
		suite.ctx,
		evmutiltypes.ModuleName,
		sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(0))),
	)

	contractAddr, err := suite.evmutilKeeper.DeployMintableERC20Contract(suite.ctx, "USDC", "USDC", uint8(18))
	suite.Require().NoError(err)
	suite.Require().Greater(len(contractAddr.Address), 0)
	return contractAddr
}

func (suite *EIP712TestSuite) TestEIP712Tx() {
	encodingConfig := app.MakeEncodingConfig()

	testcases := []struct {
		name           string
		usdcDepositAmt int64
		usdxToMintAmt  int64
		updateTx       func(txBuilder client.TxBuilder, msgs []sdk.Msg) client.TxBuilder
		updateMsgs     func(msgs []sdk.Msg) []sdk.Msg
		expectedCode   uint32
		failCheckTx    bool
		errMsg         string
	}{
		{
			name:           "processes eip712 messages successfully",
			usdcDepositAmt: 100,
			usdxToMintAmt:  99,
		},
		{
			name:           "fails when convertion more erc20 usdc than balance",
			usdcDepositAmt: 51_000,
			usdxToMintAmt:  100,
			errMsg:         "transfer amount exceeds balance",
		},
		{
			name:           "fails when minting more usdx than allowed",
			usdcDepositAmt: 100,
			usdxToMintAmt:  100,
			errMsg:         "proposed collateral ratio is below liquidation ratio",
		},
		{
			name:           "fails when trying to convert usdc for another address",
			usdcDepositAmt: 100,
			usdxToMintAmt:  90,
			errMsg:         "unauthorized",
			failCheckTx:    true,
			updateMsgs: func(msgs []sdk.Msg) []sdk.Msg {
				convertMsg := evmutiltypes.NewMsgConvertERC20ToCoin(
					suite.testEVMAddr2,
					suite.testAddr,
					suite.usdcEVMAddr,
					suite.getEVMAmount(100),
				)
				msgs[0] = &convertMsg
				return msgs
			},
		},
		{
			name:           "fails when trying to convert erc20 for non-whitelisted contract",
			usdcDepositAmt: 100,
			usdxToMintAmt:  90,
			errMsg:         "ERC20 token not enabled to convert to sdk.Coin",
			updateMsgs: func(msgs []sdk.Msg) []sdk.Msg {
				convertMsg := evmutiltypes.NewMsgConvertERC20ToCoin(
					suite.testEVMAddr,
					suite.testAddr,
					suite.testEVMAddr2,
					suite.getEVMAmount(100),
				)
				msgs[0] = &convertMsg
				return msgs
			},
		},
		{
			name:           "fails when signer tries to send messages with invalid signature",
			usdcDepositAmt: 100,
			usdxToMintAmt:  90,
			failCheckTx:    true,
			errMsg:         "tx intended signer does not match the given signer",
			updateTx: func(txBuilder client.TxBuilder, msgs []sdk.Msg) client.TxBuilder {
				var option *codectypes.Any
				option, _ = codectypes.NewAnyWithValue(&etherminttypes.ExtensionOptionsWeb3Tx{
					FeePayer:         suite.testAddr.String(),
					TypedDataChainID: 1,
					FeePayerSig:      []byte("sig"),
				})
				builder, _ := txBuilder.(authtx.ExtensionOptionsTxBuilder)
				builder.SetExtensionOptions(option)
				return txBuilder
			},
		},
		{
			name:           "fails when insufficient gas fees are provided",
			usdcDepositAmt: 100,
			usdxToMintAmt:  90,
			errMsg:         "insufficient funds",
			updateTx: func(txBuilder client.TxBuilder, msgs []sdk.Msg) client.TxBuilder {
				bk := suite.tApp.GetBankKeeper()
				gasCoins := bk.GetBalance(suite.ctx, suite.testAddr, "ukava")
				suite.tApp.GetBankKeeper().SendCoins(suite.ctx, suite.testAddr, suite.testAddr2, sdk.NewCoins(gasCoins))
				return txBuilder
			},
		},
		{
			name:           "fails when invalid chain id is provided",
			usdcDepositAmt: 100,
			usdxToMintAmt:  90,
			failCheckTx:    true,
			errMsg:         "invalid chain-id",
			updateTx: func(txBuilder client.TxBuilder, msgs []sdk.Msg) client.TxBuilder {
				gasAmt := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20)))
				return suite.createTestEIP712CosmosTxBuilder(
					suite.testAddr, suite.testPrivKey, "kavatest_12-1", uint64(helpers.DefaultGenTxGas*2), gasAmt, msgs,
				)
			},
		},
		{
			name:           "fails when invalid fee payer is provided",
			usdcDepositAmt: 100,
			usdxToMintAmt:  90,
			failCheckTx:    true,
			errMsg:         "invalid pubkey",
			updateTx: func(txBuilder client.TxBuilder, msgs []sdk.Msg) client.TxBuilder {
				gasAmt := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20)))
				return suite.createTestEIP712CosmosTxBuilder(
					suite.testAddr2, suite.testPrivKey2, ChainID, uint64(helpers.DefaultGenTxGas*2), gasAmt, msgs,
				)
			},
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// create messages to convert, mint, and deposit to lend
			usdcAmt := suite.getEVMAmount(tc.usdcDepositAmt)
			convertMsg := evmutiltypes.NewMsgConvertERC20ToCoin(
				suite.testEVMAddr,
				suite.testAddr,
				suite.usdcEVMAddr,
				usdcAmt,
			)
			usdxAmt := sdk.NewInt(1_000_000).Mul(sdk.NewInt(tc.usdxToMintAmt))
			mintMsg := cdptypes.NewMsgCreateCDP(
				suite.testAddr,
				sdk.NewCoin(USDCCoinDenom, usdcAmt),
				sdk.NewCoin(cdptypes.DefaultStableDenom, usdxAmt),
				USDCCDPType,
			)
			lendMsg := hardtypes.NewMsgDeposit(
				suite.testAddr,
				sdk.NewCoins(sdk.NewCoin(cdptypes.DefaultStableDenom, usdxAmt)),
			)
			msgs := []sdk.Msg{
				&convertMsg,
				&mintMsg,
				&lendMsg,
			}
			if tc.updateMsgs != nil {
				msgs = tc.updateMsgs(msgs)
			}

			gasAmt := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20)))
			txBuilder := suite.createTestEIP712CosmosTxBuilder(
				suite.testAddr, suite.testPrivKey, ChainID, uint64(helpers.DefaultGenTxGas*2), gasAmt, msgs,
			)
			if tc.updateTx != nil {
				txBuilder = tc.updateTx(txBuilder, msgs)
			}
			txBytes, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
			suite.Require().NoError(err)

			resCheckTx := suite.tApp.CheckTx(
				abci.RequestCheckTx{
					Tx:   txBytes,
					Type: abci.CheckTxType_New,
				},
			)
			if !tc.failCheckTx {
				suite.Require().Equal(resCheckTx.Code, uint32(0), resCheckTx.Log)
			} else {
				suite.Require().NotEqual(resCheckTx.Code, uint32(0), resCheckTx.Log)
				suite.Require().Contains(resCheckTx.Log, tc.errMsg)
			}

			resDeliverTx := suite.tApp.DeliverTx(
				abci.RequestDeliverTx{
					Tx: txBytes,
				},
			)

			if tc.errMsg == "" {
				suite.Require().Equal(resDeliverTx.Code, uint32(0), resDeliverTx.Log)

				// validate user cosmos erc20/usd balance
				bk := suite.tApp.GetBankKeeper()
				amt := bk.GetBalance(suite.ctx, suite.testAddr, USDCCoinDenom)
				suite.Require().Equal(sdk.ZeroInt(), amt.Amount)

				// validate cdp
				cdp, found := suite.tApp.GetCDPKeeper().GetCdpByOwnerAndCollateralType(suite.ctx, suite.testAddr, USDCCDPType)
				suite.Require().True(found)
				suite.Require().Equal(suite.testAddr, cdp.Owner)
				suite.Require().Equal(sdk.NewCoin(USDCCoinDenom, suite.getEVMAmount(100)), cdp.Collateral)
				suite.Require().Equal(sdk.NewCoin("usdx", sdk.NewInt(99_000_000)), cdp.Principal)

				// validate hard
				hardDeposit, found := suite.tApp.GetHardKeeper().GetDeposit(suite.ctx, suite.testAddr)
				suite.Require().True(found)
				suite.Require().Equal(suite.testAddr, hardDeposit.Depositor)
				suite.Require().Equal(sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(99_000_000))), hardDeposit.Amount)
			} else {
				suite.Require().NotEqual(resDeliverTx.Code, uint32(0), resCheckTx.Log)
				suite.Require().Contains(resDeliverTx.Log, tc.errMsg)
			}
		})
	}
}

func TestEIP712Suite(t *testing.T) {
	suite.Run(t, new(EIP712TestSuite))
}
