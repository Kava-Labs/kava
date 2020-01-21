package txs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kava-labs/kava/cmd/kvoracle/types"
)

const (
	BaseURL          = "http://localhost:1317"
	coinGeckoBaseURL = "https://api.coingecko.com/api/v3/coins/"
)

// GetCoinGeckoPrices gets prices for an array of coins by their symbols
func GetCoinGeckoPrices(symbols []string, convert string) []types.Asset {
	var assets []types.Asset
	marketCodeDict := buildMarketCodeDict()

	for _, symbol := range symbols {
		requestURL := fmt.Sprintf("%s/%s/tickers", coinGeckoBaseURL, symbol)

		resp, err := makeReq(requestURL, convert)
		if err != nil {
			fmt.Println(err)
		}

		// Unmarshal the response to a usable format
		var data *types.CoinGeckoTickers
		err = json.Unmarshal(resp, &data)
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
						TargetMarketCode: marketCodeDict[symbol],
					}
					assets = append(assets, asset)
				}
			}
		}
	}
	return assets
}

// GetAssetList gets a list of assets on kava
func GetAssetList() {
	// Format URL and HTTP request
	requestURL := fmt.Sprintf("%s/%s", BaseURL, "pricefeed/assets")
	resp, err := makeReq(requestURL, "")
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal the response to a usable format
	var data *types.MarketsRes
	err = json.Unmarshal(resp, &data)
	if err != nil {
		fmt.Println(err)
	}

	// TODO: Unmarshal to object instead of string
	for _, asset := range data.Result {
		fmt.Println(asset)
	}
}

// GetAssetPrice gets an asset's current price on kava
func GetAssetPrice(symbol string) {
	// Format URL and HTTP request
	requestURL := fmt.Sprintf("%s/%s/%s", BaseURL, "pricefeed/currentprice", symbol)
	resp, err := makeReq(requestURL, "")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(resp))
}

// makeReq HTTP request helper
func makeReq(reqURL string, convert string) ([]byte, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	// TODO: generalize this to accept any JSON param
	if convert != "" {
		q := url.Values{}
		q.Add("convert", convert)
		req.Header.Set("Accepts", "application/json")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := doReq(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// doReq HTTP client
func doReq(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if 200 != resp.StatusCode {
		return nil, fmt.Errorf("%s", body)
	}

	return body, nil
}

// TODO: Replace this with dynamically populated asset list
func buildMarketCodeDict() map[string]string {
	var codeDict = make(map[string]string)

	codeDict["bitcoin"] = "btc-usd"
	codeDict["kava"] = "kava-usd"
	codeDict["ripple"] = "xrp-usd"
	codeDict["binancecoin"] = "bnb-usd"
	codeDict["cosmos"] = "atom-usd"

	return codeDict
}
