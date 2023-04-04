package testutil

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/server/config"
	etherminttests "github.com/evmos/ethermint/tests"
	etherminttypes "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	tmtime "github.com/tendermint/tendermint/types/time"
	"github.com/tendermint/tendermint/version"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type Suite struct {
	suite.Suite

	App            app.TestApp
	Ctx            sdk.Context
	Address        common.Address
	BankKeeper     bankkeeper.Keeper
	AccountKeeper  authkeeper.AccountKeeper
	Keeper         keeper.Keeper
	EvmBankKeeper  keeper.EvmBankKeeper
	Addrs          []sdk.AccAddress
	EvmModuleAddr  sdk.AccAddress
	QueryClient    types.QueryClient
	QueryClientEvm evmtypes.QueryClient
	Key1           *ethsecp256k1.PrivKey
	Key1Addr       types.InternalEVMAddress
	Key2           *ethsecp256k1.PrivKey
}

func (suite *Suite) SetupTest() {
	tApp := app.NewTestApp()

	suite.Ctx = tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.App = tApp
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.AccountKeeper = tApp.GetAccountKeeper()
	suite.Keeper = tApp.GetEvmutilKeeper()
	suite.EvmBankKeeper = keeper.NewEvmBankKeeper(tApp.GetEvmutilKeeper(), suite.BankKeeper, suite.AccountKeeper)
	suite.EvmModuleAddr = suite.AccountKeeper.GetModuleAddress(evmtypes.ModuleName)

	// test evm user keys that have no minting permissions
	key1, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.Key1 = key1
	suite.Key1Addr = types.NewInternalEVMAddress(common.BytesToAddress(suite.Key1.PubKey().Address()))
	suite.Key2, err = ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	suite.Addrs = addrs

	evmGenesis := evmtypes.DefaultGenesisState()
	evmGenesis.Params.EvmDenom = "akava"

	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	cdc := suite.App.AppCodec()
	coins := sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000_000_000_000_000_000))
	authGS := app.NewFundedGenStateWithSameCoins(cdc, coins, []sdk.AccAddress{
		sdk.AccAddress(suite.Key1.PubKey().Address()),
		sdk.AccAddress(suite.Key2.PubKey().Address()),
	})

	gs := app.GenesisState{
		evmtypes.ModuleName:       cdc.MustMarshalJSON(evmGenesis),
		feemarkettypes.ModuleName: cdc.MustMarshalJSON(feemarketGenesis),
	}
	suite.App.InitializeFromGenesisStates(authGS, gs)

	// consensus key - needed to set up evm module
	consPriv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	consAddress := sdk.ConsAddress(consPriv.PubKey().Address())

	// InitializeFromGenesisStates commits first block so we start at 2 here
	suite.Ctx = suite.App.NewContext(false, tmproto.Header{
		Height:          suite.App.LastBlockHeight() + 1,
		ChainID:         "kavatest_1-1",
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

	// We need to set the validator as calling the EVM looks up the validator address
	// https://github.com/evmos/ethermint/blob/f21592ebfe74da7590eb42ed926dae970b2a9a3f/x/evm/keeper/state_transition.go#L487
	// evmkeeper.EVMConfig() will return error "failed to load evm config" if not set
	acc := &etherminttypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.Address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
	suite.AccountKeeper.SetAccount(suite.Ctx, acc)
	valAddr := sdk.ValAddress(suite.Address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, consPriv.PubKey(), stakingtypes.Description{})
	suite.Require().NoError(err)
	err = suite.App.GetStakingKeeper().SetValidatorByConsAddr(suite.Ctx, validator)
	suite.Require().NoError(err)
	suite.App.GetStakingKeeper().SetValidator(suite.Ctx, validator)

	// add conversion pair for first module deployed contract to evmutil params
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(
		types.NewConversionPairs(
			types.NewConversionPair(
				// First contract this module deploys
				MustNewInternalEVMAddressFromString("0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3"),
				"erc20/usdc",
			),
		),
	))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelper, suite.App.GetEvmKeeper())
	suite.QueryClientEvm = evmtypes.NewQueryClient(queryHelper)
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.App.GetEvmutilKeeper()))
	suite.QueryClient = types.NewQueryClient(queryHelper)

	// We need to commit so that the ethermint feemarket beginblock runs to set the minfee
	// feeMarketKeeper.GetBaseFee() will return nil otherwise
	suite.Commit()
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

func (suite *Suite) FundAccountWithKava(addr sdk.AccAddress, coins sdk.Coins) {
	ukava := coins.AmountOf("ukava")
	if ukava.IsPositive() {
		err := suite.App.FundAccount(suite.Ctx, addr, sdk.NewCoins(sdk.NewCoin("ukava", ukava)))
		suite.Require().NoError(err)
	}
	akava := coins.AmountOf("akava")
	if akava.IsPositive() {
		err := suite.Keeper.SetBalance(suite.Ctx, addr, akava)
		suite.Require().NoError(err)
	}
}

func (suite *Suite) FundModuleAccountWithKava(moduleName string, coins sdk.Coins) {
	ukava := coins.AmountOf("ukava")
	if ukava.IsPositive() {
		err := suite.App.FundModuleAccount(suite.Ctx, moduleName, sdk.NewCoins(sdk.NewCoin("ukava", ukava)))
		suite.Require().NoError(err)
	}
	akava := coins.AmountOf("akava")
	if akava.IsPositive() {
		addr := suite.AccountKeeper.GetModuleAddress(moduleName)
		err := suite.Keeper.SetBalance(suite.Ctx, addr, akava)
		suite.Require().NoError(err)
	}
}

func (suite *Suite) DeployERC20() types.InternalEVMAddress {
	// make sure module account is created
	// qq: any better ways to do this?
	suite.App.FundModuleAccount(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(0))),
	)

	contractAddr, err := suite.Keeper.DeployTestMintableERC20Contract(suite.Ctx, "USDC", "USDC", uint8(18))
	suite.Require().NoError(err)
	suite.Require().Greater(len(contractAddr.Address), 0)
	return contractAddr
}

func (suite *Suite) GetERC20BalanceOf(
	contractAbi abi.ABI,
	contractAddr types.InternalEVMAddress,
	accountAddr types.InternalEVMAddress,
) *big.Int {
	// Query ERC20.balanceOf()
	addr := common.BytesToAddress(suite.Key1.PubKey().Address())
	res, err := suite.QueryContract(
		types.ERC20MintableBurnableContract.ABI,
		addr,
		suite.Key1,
		contractAddr,
		"balanceOf",
		accountAddr.Address,
	)
	suite.Require().NoError(err)
	suite.Require().Len(res, 1)

	balance, ok := res[0].(*big.Int)
	suite.Require().True(ok, "balanceOf should respond with *big.Int")
	return balance
}

func (suite *Suite) QueryContract(
	contractAbi abi.ABI,
	from common.Address,
	fromKey *ethsecp256k1.PrivKey,
	contract types.InternalEVMAddress,
	method string,
	args ...interface{},
) ([]interface{}, error) {
	// Pack query args
	data, err := contractAbi.Pack(method, args...)
	suite.Require().NoError(err)

	// Send TX
	res, err := suite.SendTx(contract, from, fromKey, data)
	suite.Require().NoError(err)

	// Check for VM errors and unpack returned data
	switch res.VmError {
	case vm.ErrExecutionReverted.Error():
		response, err := abi.UnpackRevert(res.Ret)
		suite.Require().NoError(err)

		return nil, errors.New(response)
	case "": // No error, continue
	default:
		panic(fmt.Sprintf("unhandled vm error response: %v", res.VmError))
	}

	// Unpack response
	unpackedRes, err := contractAbi.Unpack(method, res.Ret)
	suite.Require().NoErrorf(err, "failed to unpack method %v response", method)

	return unpackedRes, nil
}

// SendTx submits a transaction to the block.
func (suite *Suite) SendTx(
	contractAddr types.InternalEVMAddress,
	from common.Address,
	signerKey *ethsecp256k1.PrivKey,
	transferData []byte,
) (*evmtypes.MsgEthereumTxResponse, error) {
	ctx := sdk.WrapSDKContext(suite.Ctx)
	chainID := suite.App.GetEvmKeeper().ChainID()

	args, err := json.Marshal(&evmtypes.TransactionArgs{
		To:   &contractAddr.Address,
		From: &from,
		Data: (*hexutil.Bytes)(&transferData),
	})
	if err != nil {
		return nil, err
	}
	gasRes, err := suite.QueryClientEvm.EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	if err != nil {
		return nil, err
	}

	nonce := suite.App.GetEvmKeeper().GetNonce(suite.Ctx, suite.Address)

	baseFee := suite.App.GetFeeMarketKeeper().GetBaseFee(suite.Ctx)
	suite.Require().NotNil(baseFee, "base fee is nil")

	// Mint the max gas to the FeeCollector to ensure balance in case of refund
	suite.MintFeeCollector(sdk.NewCoins(
		sdk.NewCoin(
			"ukava",
			sdk.NewInt(baseFee.Int64()*int64(gasRes.Gas*2)),
		)))

	ercTransferTx := evmtypes.NewTx(
		chainID,
		nonce,
		&contractAddr.Address,
		nil,          // amount
		gasRes.Gas*2, // gasLimit, TODO: runs out of gas with just res.Gas, ex: estimated was 21572 but used 24814
		nil,          // gasPrice
		suite.App.GetFeeMarketKeeper().GetBaseFee(suite.Ctx), // gasFeeCap
		big.NewInt(1), // gasTipCap
		transferData,
		&ethtypes.AccessList{}, // accesses
	)

	ercTransferTx.From = hex.EncodeToString(signerKey.PubKey().Address())
	err = ercTransferTx.Sign(ethtypes.LatestSignerForChainID(chainID), etherminttests.NewSigner(signerKey))
	if err != nil {
		return nil, err
	}

	rsp, err := suite.App.GetEvmKeeper().EthereumTx(ctx, ercTransferTx)
	if err != nil {
		return nil, err
	}
	// Do not check vm error here since we want to check for errors later

	return rsp, nil
}

func (suite *Suite) MintFeeCollector(coins sdk.Coins) {
	err := suite.App.FundModuleAccount(suite.Ctx, authtypes.FeeCollectorName, coins)
	suite.Require().NoError(err)
}

// GetEvents returns emitted events on the sdk context
func (suite *Suite) GetEvents() sdk.Events {
	return suite.Ctx.EventManager().Events()
}

// EventsContains asserts that the expected event is in the provided events
func (suite *Suite) EventsContains(events sdk.Events, expectedEvent sdk.Event) {
	foundMatch := false
	for _, event := range events {
		if event.Type == expectedEvent.Type {
			if reflect.DeepEqual(attrsToMap(expectedEvent.Attributes), attrsToMap(event.Attributes)) {
				foundMatch = true
			}
		}
	}

	suite.Truef(foundMatch, "event of type %s not found or did not match", expectedEvent.Type)
}

// EventsDoNotContain asserts that the event is **not** is in the provided events
func (suite *Suite) EventsDoNotContain(events sdk.Events, eventType string) {
	foundMatch := false
	for _, event := range events {
		if event.Type == eventType {
			foundMatch = true
		}
	}

	suite.Falsef(foundMatch, "event of type %s should not be found, but was found", eventType)
}

func attrsToMap(attrs []abci.EventAttribute) []sdk.Attribute {
	out := []sdk.Attribute{}

	for _, attr := range attrs {
		out = append(out, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
	}

	return out
}

// MustNewInternalEVMAddressFromString returns a new InternalEVMAddress from a
// hex string. This will panic if the input hex string is invalid.
func MustNewInternalEVMAddressFromString(addrStr string) types.InternalEVMAddress {
	addr, err := types.NewInternalEVMAddressFromString(addrStr)
	if err != nil {
		panic(err)
	}

	return addr
}
