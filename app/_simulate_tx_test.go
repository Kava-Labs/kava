package app_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kava-labs/kava/app"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	jsonrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

type SimulateRequestTestSuite struct {
	suite.Suite
	cliCtx           context.CLIContext
	restServer       *httptest.Server
	rpcServer        *httptest.Server
	simulateResponse func(jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse
}

func (suite *SimulateRequestTestSuite) SetupTest() {
	suite.rpcServer = rpcTestServer(suite.T(), func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		suite.Require().Equal("abci_query", request.Method)
		return suite.simulateResponse(request)
	})
	cdc := app.MakeCodec()
	suite.cliCtx = context.CLIContext{}.WithCodec(cdc).WithNodeURI(suite.rpcServer.URL)

	router := mux.NewRouter()
	app.RegisterSimulateRoutes(suite.cliCtx, router)
	suite.restServer = httptest.NewServer(router)
}

func (suite *SimulateRequestTestSuite) TearDownTest() {
	suite.rpcServer.Close()
	suite.restServer.Close()
}

func (suite *SimulateRequestTestSuite) TestSimulateRequest() {
	fromAddr, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	suite.Require().NoError(err)
	toAddr, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	suite.Require().NoError(err)

	simRequest := app.SimulateRequest{
		Msgs: []sdk.Msg{
			bank.MsgSend{
				FromAddress: fromAddr,
				ToAddress:   toAddr,
				Amount:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
			},
		},
		Fee: auth.StdFee{
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(5e4))),
			Gas:    1e6,
		},
		Memo: "test memo",
	}
	requestBody, err := suite.cliCtx.Codec.MarshalJSON(simRequest)
	suite.Require().NoError(err)

	mockResponse := sdk.SimulationResponse{
		GasInfo: sdk.GasInfo{
			GasWanted: 500000,
			GasUsed:   200000,
		},
	}
	suite.simulateResponse = func(rpcRequest jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		var params struct {
			Path   string
			Data   tmbytes.HexBytes
			Height string
			Prove  bool
		}

		err := json.Unmarshal(rpcRequest.Params, &params)
		suite.Require().NoError(err)
		suite.Require().Equal("0", params.Height)

		var tx auth.StdTx
		err = suite.cliCtx.Codec.UnmarshalBinaryLengthPrefixed(params.Data, &tx)
		suite.Require().NoError(err)

		// assert tx is generated and passed correctly from the simulate request
		suite.Equal(simRequest.Msgs, tx.Msgs)
		suite.Equal(simRequest.Fee, tx.Fee)
		suite.Equal([]auth.StdSignature{{}}, tx.Signatures)
		suite.Equal(simRequest.Memo, tx.Memo)

		respValue, err := suite.cliCtx.Codec.MarshalBinaryBare(mockResponse)
		suite.Require().NoError(err)

		abciResult := ctypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				Height: 100000,
				Value:  respValue,
			},
		}

		data, err := suite.cliCtx.Codec.MarshalJSON(&abciResult)
		suite.Require().NoError(err)

		return jsonrpctypes.RPCResponse{
			JSONRPC: rpcRequest.JSONRPC,
			ID:      rpcRequest.ID,
			Result:  json.RawMessage(data),
		}
	}

	req, err := http.NewRequest("POST", suite.restServer.URL+"/tx/simulate", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var respWithHeight rest.ResponseWithHeight
	err = suite.cliCtx.Codec.UnmarshalJSON(body, &respWithHeight)
	suite.Require().NoError(err)

	suite.Equal(int64(100000), respWithHeight.Height)

	var simResp sdk.SimulationResponse
	err = suite.cliCtx.Codec.UnmarshalJSON(respWithHeight.Result, &simResp)
	suite.Require().NoError(err)

	suite.Equal(mockResponse, simResp)
}

func TestSimulateRequestTestSuite(t *testing.T) {
	suite.Run(t, new(SimulateRequestTestSuite))
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
