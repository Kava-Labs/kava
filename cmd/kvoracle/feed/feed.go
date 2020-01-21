package feed

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/matryer/try"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/cmd/kvoracle/txs"
	"github.com/kava-labs/kava/cmd/kvoracle/types"
)

const (
	coinGeckoBaseURL = "https://api.coingecko.com/api/v3/coins/"
)

var codeDict map[string]string

// PostPrices gets the current coin prices and posts them to kava
func PostPrices(
	coins []string,
	accAddress sdk.AccAddress,
	chainID string,
	cdc *codec.Codec,
	oracleName string,
	passphrase string,
	cliCtx context.CLIContext,
	rpcURL string,
) error {
	// Get and display current time
	now := time.Now().Format("15:04:05")
	fmt.Println("Time: ", now)

	// Get asset prices
	assets := getCoinGeckoPrices(coins, "USD")

	for i := 0; i < len(assets); i++ {
		startPostPriceRoutine(assets[i], accAddress, chainID, cdc, oracleName, passphrase, cliCtx, rpcURL)
	}

	return nil
}

// startPostPriceRoutine attempts to send MsgPostPrice. If unsuccessful due to local errors, it will
// attempt again for a total of 3 attempts. If the tx is received by the blockchain but unsuccessful
// due to blockchain state, it will not try to resend the tx - but will print the tx log text.
func startPostPriceRoutine(
	asset types.Asset,
	accAddress sdk.AccAddress,
	chainID string,
	cdc *codec.Codec,
	oracleName string,
	passphrase string,
	cliCtx context.CLIContext,
	rpcURL string,
) {

	var txRes sdk.TxResponse
	err := try.Do(func(attempt int) (bool, error) {
		var err error

		attemptStr := ""
		if attempt > 1 {
			attemptStr = fmt.Sprintf(" [attempt #%d]", attempt)
		}

		fmt.Printf("Posting price '%f' for %s...%s\n", asset.Price, asset.Symbol, attemptStr)
		txRes, err = postPriceToKava(asset, accAddress, chainID, cdc, oracleName, passphrase, cliCtx, rpcURL)
		if err != nil {
			time.Sleep(5 * time.Second)
		}
		return attempt < 3, err // try 3 times
	})

	if err != nil {
		log.Fatalln("Error:", err)
	}

	fmt.Printf("Tx hash: %s\n", txRes.TxHash)
	fmt.Printf("Tx log: %v\n\n", txRes.RawLog)
}

func postPriceToKava(
	asset types.Asset,
	accAddress sdk.AccAddress,
	chainID string,
	cdc *codec.Codec,
	oracleName string,
	passphrase string,
	cliCtx context.CLIContext,
	rpcURL string,
) (sdk.TxResponse, error) {
	// Parse the price
	price, err := sdk.NewDecFromStr(fmt.Sprintf("%f", asset.Price))
	if err != nil {
		return sdk.TxResponse{}, err
	}

	// Build the msg
	msgPostPrice, sdkErr := txs.ConstructMsgPostPrice(accAddress, price, asset.TargetMarketCode)
	if sdkErr != nil {
		return sdk.TxResponse{}, sdkErr
	}

	// Send tx containing msg to kava
	txRes, sdkErr := txs.SendTxPostPrice(chainID, cdc, accAddress, oracleName, passphrase, cliCtx, &msgPostPrice, rpcURL)
	if sdkErr != nil {
		return sdk.TxResponse{}, sdkErr
	}

	return txRes, nil
}

// getCoinGeckoPrices gets prices for an array of coins by their symbols
func getCoinGeckoPrices(symbols []string, convert string) []types.Asset {
	var assets []types.Asset
	client := &http.Client{}

	setupMarketCodeDict()

	for _, symbol := range symbols {
		// Format URL and HTTP request
		coinURL := fmt.Sprintf("%s/%s/tickers", coinGeckoBaseURL, symbol)
		req, err := http.NewRequest("GET", coinURL, nil)
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		q := url.Values{}
		q.Add("convert", convert)
		req.Header.Set("Accepts", "application/json")
		req.URL.RawQuery = q.Encode()

		// Make an HTTP request
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request to server")
			os.Exit(1)
		}

		// Read the response
		respBody, _ := ioutil.ReadAll(resp.Body)

		// Unmarshal the response to a usable format
		var data *types.CoinGeckoTickers
		err = json.Unmarshal(respBody, &data)
		if err != nil {
			fmt.Println(err)
		}

		// Use coin's USDT market from Binance
		if data != nil && data.Tickers != nil {
			for _, ticker := range data.Tickers {
				if ticker.Market.Name == "Binance" && ticker.Target == "USDT" {
					asset := types.Asset{
						Symbol:           data.Name,
						Price:            ticker.Last,
						TargetMarketCode: codeDict[symbol],
					}
					assets = append(assets, asset)
				}
			}
		}
	}
	return assets
}

// TODO: Replace this with dynamically populated asset list
//		 once resp is formatted correctly
func setupMarketCodeDict() {
	codeDict = make(map[string]string)

	// Populate the dictionary
	codeDict["bitcoin"] = "btc:usd"
	codeDict["kava"] = "kava:usd"
	codeDict["ripple"] = "xrp:usd"
	codeDict["binancecoin"] = "bnb:usd"
	codeDict["cosmos"] = "atom:usd"
}
