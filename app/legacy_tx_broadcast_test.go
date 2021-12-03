package app_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kava-labs/kava/app"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/tests/mocks"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	jsonrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

type LegacyTxBroadcastTestSuite struct {
	suite.Suite
	clientCtx        client.Context
	restServer       *httptest.Server
	rpcServer        *httptest.Server
	ctrl             *gomock.Controller
	accountRetriever *mocks.MockAccountRetriever
	simulateResponse func(jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse
}

func (suite *LegacyTxBroadcastTestSuite) SetupTest() {
	app.SetSDKConfig()

	// setup the mock rpc server
	suite.rpcServer = rpcTestServer(suite.T(), func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		suite.Require().Equal("broadcast_tx_sync", request.Method)
		return suite.simulateResponse(request)
	})

	// setup client context with rpc client, account retriever mock, codecs, and tx config
	rpcClient, err := rpchttp.New(suite.rpcServer.URL, "/websocket")
	suite.Require().NoError(err)
	suite.ctrl = gomock.NewController(suite.T())
	suite.accountRetriever = mocks.NewMockAccountRetriever(suite.ctrl)
	encodingConfig := app.MakeEncodingConfig()
	suite.clientCtx = client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithNodeURI(suite.rpcServer.URL).
		WithClient(rpcClient).
		WithAccountRetriever(suite.accountRetriever)

	// setup rest server
	router := mux.NewRouter()
	app.RegisterLegacyTxRoutes(suite.clientCtx, router)
	suite.restServer = httptest.NewServer(router)
}

func (suite *LegacyTxBroadcastTestSuite) TearDownTest() {
	suite.rpcServer.Close()
	suite.restServer.Close()
	suite.ctrl.Finish()
}

func (suite *LegacyTxBroadcastTestSuite) TestSimulateRequest() {
	_, pk, fromAddr := testdata.KeyTestPubAddr()
	toAddr, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	suite.Require().NoError(err)

	// build a legacy transaction
	msgs := []sdk.Msg{banktypes.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))))}
	fee := legacytx.NewStdFee(1e6, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(5e4))))
	sigs := []legacytx.StdSignature{legacytx.NewStdSignature(pk, []byte("an amino json signed signature"))}
	stdTx := legacytx.NewStdTx(msgs, fee, sigs, "legacy broadcast test")
	stdTx.TimeoutHeight = 100000
	txReq := app.LegacyTxBroadcastRequest{
		Tx:   stdTx,
		Mode: "sync",
	}

	// setup mock tendermint jsonrpc handler for BroadcastTx
	var broadcastedTx authsigning.Tx
	var broadcastedTxHash tmbytes.HexBytes
	suite.simulateResponse = func(rpcRequest jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		var params struct {
			Tx tmtypes.Tx `json:"tx"`
		}

		err := json.Unmarshal(rpcRequest.Params, &params)
		suite.Require().NoError(err)
		decodedTx, err := suite.clientCtx.TxConfig.TxDecoder()(params.Tx)
		suite.Require().NoError(err)
		wrappedTx, err := suite.clientCtx.TxConfig.WrapTxBuilder(decodedTx)
		suite.Require().NoError(err)

		broadcastedTx = wrappedTx.GetTx()
		broadcastedTxHash = params.Tx.Hash()

		resp := &ctypes.ResultBroadcastTx{
			Log:  "[]",
			Hash: broadcastedTxHash,
		}
		result, err := suite.clientCtx.LegacyAmino.MarshalJSON(resp)
		suite.Require().NoError(err)

		return jsonrpctypes.RPCResponse{
			JSONRPC: rpcRequest.JSONRPC,
			ID:      rpcRequest.ID,
			Result:  json.RawMessage(result),
		}
	}

	// mock account sequence retrieval
	suite.accountRetriever.EXPECT().
		GetAccountNumberSequence(suite.clientCtx, fromAddr).
		Return(uint64(100), uint64(101), nil)

	// amino encode legacy tx
	requestBody, err := suite.clientCtx.LegacyAmino.MarshalJSON(txReq)
	suite.Require().NoError(err)

	// post transaction to POST /txs
	req, err := http.NewRequest("POST", suite.restServer.URL+"/txs", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	// assert broadcasted tx has all information
	suite.Equal(msgs, broadcastedTx.GetMsgs())
	suite.Equal(fee.Amount, broadcastedTx.GetFee())
	suite.Equal(fee.Gas, broadcastedTx.GetGas())
	suite.Equal(stdTx.TimeoutHeight, broadcastedTx.GetTimeoutHeight())
	suite.Equal(stdTx.Memo, broadcastedTx.GetMemo())

	// assert broadcasted tx has correct signature
	stdSignatures, err := stdTx.GetSignaturesV2()
	suite.Require().NoError(err)
	stdSignatures[0].Sequence = uint64(101)
	broadcastedSigs, err := broadcastedTx.GetSignaturesV2()
	suite.Require().NoError(err)
	suite.Equal(stdSignatures, broadcastedSigs)

	// decode response body
	body, err := ioutil.ReadAll(resp.Body)
	suite.Require().NoError(err)
	var txResponse sdk.TxResponse
	err = suite.clientCtx.LegacyAmino.UnmarshalJSON(body, &txResponse)
	suite.Require().NoError(err)

	// ensure response is correct
	suite.Equal("[]", txResponse.RawLog)
	suite.Equal(broadcastedTxHash.String(), txResponse.TxHash)
}

func TestLegacyTxBroadcastTestSuite(t *testing.T) {
	suite.Run(t, new(LegacyTxBroadcastTestSuite))
}

func rpcTestServer(
	t *testing.T,
	rpcHandler func(jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse,
) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var request jsonrpctypes.RPCRequest
		err = json.Unmarshal(body, &request)
		require.NoError(t, err)

		response := rpcHandler(request)

		b, err := json.Marshal(&response)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}))
}
