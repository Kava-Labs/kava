package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/require"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttpclient "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	rpcUrl        = "http://localhost:26657"
	faucetAccount = "kava1adkm6svtzjsxxvg7g6rshg6kj9qwej8gwqadqd"
)

func TestLatestHeightBalanceQuery(t *testing.T) {
	t.Skip()
	app.SetSDKConfig()
	encodingConfig := app.MakeEncodingConfig()
	cdc := encodingConfig.Amino

	client, err := rpchttpclient.New(rpcUrl, "/websocket")
	require.NoError(t, err)

	jobCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	numJobs := 100
	errChan := make(chan error, numJobs)

	for i := 0; i < numJobs; i++ {
		go func() {
			// Start at random times
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			for jobCtx.Err() == nil {
				coins, err := getLatestBalance(client, cdc, faucetAccount)
				if err != nil {
					errChan <- err
					return
				}

				if len(coins) == 0 {
					errChan <- errors.New("received zero coins")
				}
			}
		}()
	}

	select {
	case err := <-errChan:
		require.NoError(t, err)
	case <-jobCtx.Done():
	}
}

func getLatestBalance(client rpcclient.Client, cdc *codec.LegacyAmino, address string) (sdk.Coins, error) {
	addr, err := sdk.AccAddressFromBech32(faucetAccount)
	if err != nil {
		return nil, err
	}

	bz, err := cdc.MarshalJSON(banktypes.NewQueryAllBalancesRequest(addr, nil))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", banktypes.QuerierRoute, banktypes.QueryAllBalances)
	opts := rpcclient.ABCIQueryOptions{Prove: false}

	result, err := parseABCIResult(client.ABCIQueryWithOptions(context.Background(), path, bz, opts))
	if err != nil {
		return nil, err
	}

	var balance sdk.Coins
	err = cdc.UnmarshalJSON(result, &balance)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func parseABCIResult(result *ctypes.ResultABCIQuery, err error) ([]byte, error) {
	if err != nil {
		return []byte{}, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return []byte{}, errors.New(resp.Log)
	}

	value := result.Response.GetValue()
	if value == nil {
		return []byte{}, nil
	}

	return value, nil
}
